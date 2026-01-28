package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gobot/app"
	"gobot/internal/config"
	"gobot/internal/handler"
	"gobot/internal/mcp"
	mcpoauth "gobot/internal/mcp/oauth"
	"gobot/internal/channels"
	"gobot/internal/middleware"
	"gobot/internal/oauth"
	"gobot/internal/realtime"
	"gobot/internal/router"
	"gobot/internal/svc"
	"gobot/internal/websocket"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"golang.org/x/crypto/acme/autocert"
)

//go:embed etc/gobot.yaml
var embeddedConfig []byte

var _ = flag.String("f", "etc/gobot.yaml", "the config file (ignored, using embedded config)")

func main() {
	flag.Parse()

	var c config.Config

	if err := conf.LoadFromYamlBytes([]byte(os.ExpandEnv(string(embeddedConfig))), &c); err != nil {
		fmt.Printf("Failed to load embedded config: %v\n", err)
		os.Exit(1)
	}

	var srvHost = c.Host
	var serverPort = c.Port
	var useHTTPS = false

	if c.IsProductionMode() {
		if c.App.Domain == "" {
			fmt.Println("ERROR: App.Domain is required in production mode")
			os.Exit(1)
		}
		srvHost = c.App.Domain
		serverPort = 443
		useHTTPS = true
		fmt.Printf("Running in PRODUCTION mode - server.json will return https://%s\n", c.App.Domain)
	} else if serverPort == 443 || serverPort == 80 {
		if c.App.Domain == "" {
			fmt.Println("ERROR: App.Domain is required when using ports 80/443")
			os.Exit(1)
		}
		srvHost = c.App.Domain
		if serverPort == 443 {
			useHTTPS = true
		}
	} else {
		srvHost = "localhost"
		app.DevMode = true
		fmt.Printf("Running in DEVELOPMENT mode - server.json will return http://localhost:%d\n", serverPort)
	}

	fmt.Println("Server Host:", srvHost, "Port:", serverPort, "Use HTTPS:", useHTTPS)

	app.SetServerHost(srvHost, serverPort, useHTTPS)

	spaFS, err := app.FileSystem()
	if err != nil {
		fmt.Printf("Warning: Could not load embedded SPA files: %v\n", err)
		fmt.Println("App should be running separately on port 5173")
	}

	var serverOpts []rest.RunOption
	if err == nil {
		serverOpts = append(serverOpts,
			rest.WithNotFoundHandler(app.NotFoundHandler(spaFS)),
		)
	}
	server := rest.MustNewServer(c.RestConf, serverOpts...)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if c.IsSecurityHeadersEnabled() {
				headers := middleware.APISecurityHeaders()
				w.Header().Set("Content-Security-Policy", headers.ContentSecurityPolicy)
				w.Header().Set("X-Content-Type-Options", headers.XContentTypeOptions)
				w.Header().Set("X-Frame-Options", headers.XFrameOptions)
				w.Header().Set("X-XSS-Protection", headers.XXSSProtection)
				w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)
				w.Header().Set("Permissions-Policy", headers.PermissionsPolicy)
				if c.IsForceHTTPS() {
					w.Header().Set("Strict-Transport-Security", headers.StrictTransportSecurity)
				}
				w.Header().Set("Cache-Control", headers.CacheControl)
				w.Header().Set("Pragma", headers.Pragma)
			}
			next(w, r)
		}
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/csrf-token",
		Handler: ctx.SecurityMiddleware.GetCSRFTokenHandler(),
	})

	handler.RegisterHandlers(server, ctx)

	if ctx.UseLocal() && c.IsOAuthEnabled() {
		oauthHandler := oauth.NewHandler(ctx)
		oauthHandler.RegisterRoutes(http.DefaultServeMux)
		fmt.Println("OAuth callbacks registered at /oauth/{provider}/callback")
	}

	if ctx.UseLocal() {
		var baseURL string
		if useHTTPS {
			baseURL = fmt.Sprintf("https://%s", srvHost)
		} else {
			baseURL = fmt.Sprintf("http://%s:%d", srvHost, serverPort)
		}

		mcpHandler := mcp.NewHandler(ctx, baseURL)
		http.DefaultServeMux.Handle("/mcp", mcpHandler)
		http.DefaultServeMux.Handle("/mcp/", mcpHandler)

		mcpOAuthHandler := mcpoauth.NewHandler(ctx, baseURL)
		mcpOAuthHandler.RegisterRoutes(http.DefaultServeMux)

		fmt.Println("MCP endpoint registered at /mcp")
		fmt.Println("MCP OAuth endpoints registered at /.well-known/oauth-* and /mcp/oauth/*")
	}

	hub := realtime.NewHub()
	go hub.Run(context.Background())

	go ctx.AgentHub.Run(context.Background())

	// Initialize message router for channel → agent routing
	channelMgr := channels.NewManager()
	msgRouter := router.NewRouter(channelMgr, ctx.AgentHub)
	_ = msgRouter // Router is ready for use when channels are configured

	rewriteHandler := realtime.NewRewriteHandler(ctx)
	rewriteHandler.Register()

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws",
		Handler: websocket.Handler(hub),
	})

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/agent/ws",
		Handler: agentWebSocketHandler(ctx),
	})

	if app.DevMode {
		fmt.Printf("Starting go-zero backend server on %s:%d (dev mode)...\n", c.Host, c.Port)
		server.Start()
		return
	}

	go func() {
		fmt.Printf("Starting go-zero backend server on %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(c.App.Domain, "www."+c.App.Domain),
		Email:      c.App.AdminEmail,
	}

	backendURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", c.Host, c.Port))
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if req.Header.Get("Upgrade") != "" {
			req.Header.Set("Connection", "Upgrade")
		}
	}

	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
		WriteBufferSize:     32 << 10,
		ReadBufferSize:      32 << 10,
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("Proxy error: %v\n", err)
		http.Error(w, "Backend temporarily unavailable", http.StatusBadGateway)
	}

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			certManager.HTTPHandler(nil).ServeHTTP(w, r)
			return
		}
		host, _ := strings.CutPrefix(r.Host, "www.")
		newURL := fmt.Sprintf("https://%s%s", host, r.RequestURI)
		http.Redirect(w, r, newURL, http.StatusMovedPermanently)
	})

	baseHTTPSHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if nonWWWHost, hadPrefix := strings.CutPrefix(r.Host, "www."); hadPrefix {
			newURL := fmt.Sprintf("https://%s%s", nonWWWHost, r.RequestURI)
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/webhooks/") {
			proxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/oauth/") {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/mcp") || strings.HasPrefix(r.URL.Path, "/.well-known/oauth-") {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/ws") || strings.HasPrefix(r.URL.Path, "/api/v1/agent/ws") {
			proxyWebSocket(w, r, c.Host, c.Port)
			return
		}

		if err == nil {
			app.SPAHandler(spaFS).ServeHTTP(w, r)
		} else {
			http.Error(w, "SPA not available", http.StatusServiceUnavailable)
		}
	})

	httpsHandler := middleware.Gzip(middleware.CacheControl(baseHTTPSHandler))

	httpServer := &http.Server{
		Addr:         ":80",
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Println("Starting HTTP server on :80 for ACME challenges and HTTPS redirect...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: httpsHandler,
		TLSConfig: &tls.Config{
			GetCertificate:           certManager.GetCertificate,
			MinVersion:               tls.VersionTLS12,
			NextProtos:               []string{"h2", "http/1.1", "acme-tls/1"},
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
		},
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Starting HTTPS server on :443 with Let's Encrypt auto-certificate...")
	fmt.Println("Auto-redirect: www → non-www")
	fmt.Println("Auto-redirect: HTTP → HTTPS")
	fmt.Println("API routes: /api/* (proxied to backend)")
	fmt.Println("Static SPA: /* (served directly from embedded FS)")
	fmt.Println("Optimizations: HTTP/2, connection pooling, compression")

	go func() {
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTPS server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down servers gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpsServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("HTTPS server forced to shutdown: %v\n", err)
	} else {
		fmt.Println("HTTPS server stopped")
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("HTTP server forced to shutdown: %v\n", err)
	} else {
		fmt.Println("HTTP server stopped")
	}

	fmt.Println("All servers shut down successfully")
}

func agentWebSocketHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agentID := r.URL.Query().Get("agent_id")
		orgID := r.URL.Query().Get("org_id")
		token := r.URL.Query().Get("token")

		if agentID == "" || orgID == "" || token == "" {
			http.Error(w, "Missing agent_id, org_id, or token", http.StatusBadRequest)
			return
		}

		claims, err := middleware.ValidateJWT(token, ctx.Config.Auth.AccessSecret)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID, _ := claims["userId"].(string)
		if userID == "" {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx.AgentHub.HandleWebSocket(w, r, orgID, agentID, userID)
	}
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, backendHost string, backendPort int) {
	fmt.Printf("[WS Proxy] Incoming WebSocket request: %s\n", r.URL.String())

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		fmt.Println("[WS Proxy] ERROR: ResponseWriter does not support Hijack")
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	backendAddr := fmt.Sprintf("%s:%d", backendHost, backendPort)
	fmt.Printf("[WS Proxy] Dialing backend: %s\n", backendAddr)
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		fmt.Printf("[WS Proxy] ERROR: Failed to dial backend: %v\n", err)
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		fmt.Printf("[WS Proxy] ERROR: Hijack failed: %v\n", err)
		http.Error(w, "Hijack failed", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()
	fmt.Println("[WS Proxy] Connection hijacked successfully")

	if err := r.Write(backendConn); err != nil {
		fmt.Printf("[WS Proxy] ERROR: Failed to forward request: %v\n", err)
		return
	}
	fmt.Println("[WS Proxy] Request forwarded to backend")

	if clientBuf.Reader.Buffered() > 0 {
		buffered := make([]byte, clientBuf.Reader.Buffered())
		clientBuf.Read(buffered)
		backendConn.Write(buffered)
	}

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(backendConn, clientConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, backendConn)
		done <- struct{}{}
	}()
	<-done
	fmt.Println("[WS Proxy] Connection closed")
}
