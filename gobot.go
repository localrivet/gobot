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
	"gobot/internal/middleware"
	"gobot/internal/oauth"
	"gobot/internal/realtime"
	"gobot/internal/svc"
	"gobot/internal/webhook"
	"gobot/internal/websocket"

	levee "github.com/almatuck/levee-go"
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

	// Use embedded config - expand environment variables
	if err := conf.LoadFromYamlBytes([]byte(os.ExpandEnv(string(embeddedConfig))), &c); err != nil {
		fmt.Printf("Failed to load embedded config: %v\n", err)
		os.Exit(1)
	}

	// Determine server host based on config
	var srvHost = c.Host
	var serverPort = c.Port
	var useHTTPS = false

	if c.IsProductionMode() {
		if c.App.Domain == "" {
			fmt.Println("ERROR: App.Domain is required in production mode")
			os.Exit(1)
		}
		// Production mode - use domain name with HTTPS on standard port
		srvHost = c.App.Domain
		serverPort = 443
		useHTTPS = true
		fmt.Printf("Running in PRODUCTION mode - server.json will return https://%s\n", c.App.Domain)
	} else if serverPort == 443 || serverPort == 80 {
		// Fallback: check if running on standard HTTPS/HTTP ports
		if c.App.Domain == "" {
			fmt.Println("ERROR: App.Domain is required when using ports 80/443")
			os.Exit(1)
		}
		srvHost = c.App.Domain
		if serverPort == 443 {
			useHTTPS = true
		}
	} else {
		// Development mode - use localhost with port
		srvHost = "localhost"
		app.DevMode = true
		fmt.Printf("Running in DEVELOPMENT mode - server.json will return http://localhost:%d\n", serverPort)
	}

	fmt.Println("Server Host:", srvHost, "Port:", serverPort, "Use HTTPS:", useHTTPS)

	// Set server host for server.json
	app.SetServerHost(srvHost, serverPort, useHTTPS)

	// Set up SPA filesystem for static file serving
	spaFS, err := app.FileSystem()
	if err != nil {
		fmt.Printf("Warning: Could not load embedded SPA files: %v\n", err)
		fmt.Println("App should be running separately on port 5173")
	}

	// Create go-zero server with embedded app
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

	// Apply global security middleware
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

	// Add CSRF token endpoint
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/api/v1/csrf-token",
		Handler: ctx.SecurityMiddleware.GetCSRFTokenHandler(),
	})

	handler.RegisterHandlers(server, ctx)

	// Register Stripe webhook for standalone mode (Levee has its own webhook handlers)
	if ctx.UseLocal() && ctx.Config.Stripe.WebhookSecret != "" {
		server.AddRoute(rest.Route{
			Method:  http.MethodPost,
			Path:    "/api/webhook/stripe",
			Handler: webhook.StripeHandler(ctx),
		})
		fmt.Println("Stripe webhook registered at /api/webhook/stripe")

		// Sync products to Stripe on startup
		if ctx.Billing != nil && len(ctx.Config.Products) > 0 {
			syncCtx := context.Background()
			syncedProducts, err := ctx.Billing.SyncProductsToStripe(syncCtx)
			if err != nil {
				fmt.Printf("Warning: Failed to sync products to Stripe: %v\n", err)
			} else {
				fmt.Printf("Synced %d products to Stripe\n", len(syncedProducts))
				// Update config with synced Stripe IDs for lookup
				ctx.Config.Products = syncedProducts
			}
		}
	}

	// Register Levee embedded handlers on default mux (reverse proxy routes to these)
	// This includes: email tracking, unsubscribe, confirm, and webhook endpoints
	if ctx.Levee != nil {
		ctx.Levee.RegisterHandlers(http.DefaultServeMux, "",
			levee.WithUnsubscribeRedirect("/unsubscribed"),
			levee.WithConfirmRedirect("/welcome"),
			levee.WithConfirmExpiredRedirect("/confirm-expired"),
		)
	}

	// Register OAuth callback handlers directly (bypasses go-zero for browser redirects)
	if ctx.UseLocal() && c.IsOAuthEnabled() {
		oauthHandler := oauth.NewHandler(ctx)
		oauthHandler.RegisterRoutes(http.DefaultServeMux)
		fmt.Println("OAuth callbacks registered at /oauth/{provider}/callback")
	}

	// Register MCP (Model Context Protocol) handler for AI agent access
	if ctx.UseLocal() {
		// Determine base URL for MCP OAuth discovery
		var baseURL string
		if useHTTPS {
			baseURL = fmt.Sprintf("https://%s", srvHost)
		} else {
			baseURL = fmt.Sprintf("http://%s:%d", srvHost, serverPort)
		}

		mcpHandler := mcp.NewHandler(ctx, baseURL)
		http.DefaultServeMux.Handle("/mcp", mcpHandler)
		http.DefaultServeMux.Handle("/mcp/", mcpHandler)

		// Register MCP OAuth endpoints for Dynamic Client Registration
		mcpOAuthHandler := mcpoauth.NewHandler(ctx, baseURL)
		mcpOAuthHandler.RegisterRoutes(http.DefaultServeMux)

		fmt.Println("MCP endpoint registered at /mcp")
		fmt.Println("MCP OAuth endpoints registered at /.well-known/oauth-* and /mcp/oauth/*")
	}

	// Create WebSocket hub for real-time events
	hub := realtime.NewHub()
	go hub.Run(context.Background())

	// Register rewrite handler for WebSocket messages
	rewriteHandler := realtime.NewRewriteHandler(ctx)
	rewriteHandler.Register()

	// Add WebSocket endpoint
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws",
		Handler: websocket.Handler(hub),
	})

	// In development mode, just run go-zero server directly
	if app.DevMode {
		fmt.Printf("Starting go-zero backend server on %s:%d (dev mode)...\n", c.Host, c.Port)
		server.Start()
		return
	}

	// Production mode: Start go-zero server in background, then HTTPS/HTTP servers
	go func() {
		fmt.Printf("Starting go-zero backend server on %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

	// Set up autocert for Let's Encrypt - update these for your domain
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(c.App.Domain, "www."+c.App.Domain),
		Email:      c.App.AdminEmail,
	}

	// Create reverse proxy to go-zero backend with connection pooling
	backendURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", c.Host, c.Port))
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Modify director to preserve WebSocket headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if req.Header.Get("Upgrade") != "" {
			req.Header.Set("Connection", "Upgrade")
		}
	}

	// Configure transport for optimal performance
	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
		WriteBufferSize:     32 << 10,
		ReadBufferSize:      32 << 10,
	}

	// Add error handler for backend failures
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("Proxy error: %v\n", err)
		http.Error(w, "Backend temporarily unavailable", http.StatusBadGateway)
	}

	// HTTP handler for port 80 - ACME challenges and HTTPS redirect
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			certManager.HTTPHandler(nil).ServeHTTP(w, r)
			return
		}
		host, _ := strings.CutPrefix(r.Host, "www.")
		newURL := fmt.Sprintf("https://%s%s", host, r.RequestURI)
		http.Redirect(w, r, newURL, http.StatusMovedPermanently)
	})

	// HTTPS handler for port 443
	baseHTTPSHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// www → non-www redirect
		if nonWWWHost, hadPrefix := strings.CutPrefix(r.Host, "www."); hadPrefix {
			newURL := fmt.Sprintf("https://%s%s", nonWWWHost, r.RequestURI)
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
			return
		}

		// Route API requests to backend
		if strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}

		// Route Levee webhook paths to default mux (stripe, ses)
		if r.URL.Path == "/webhooks/stripe" || r.URL.Path == "/webhooks/ses" {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		// Route other webhook requests to backend (e.g., /webhooks/levee)
		if strings.HasPrefix(r.URL.Path, "/webhooks/") {
			proxy.ServeHTTP(w, r)
			return
		}

		// Route Levee email tracking/confirmation paths to default mux
		if strings.HasPrefix(r.URL.Path, "/e/") || r.URL.Path == "/confirm-email" {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		// Route OAuth callbacks to default mux (browser redirects, not API calls)
		if strings.HasPrefix(r.URL.Path, "/oauth/") {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		// Route MCP (Model Context Protocol) requests to default mux
		if strings.HasPrefix(r.URL.Path, "/mcp") || strings.HasPrefix(r.URL.Path, "/.well-known/oauth-") {
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}

		// Route WebSocket requests to backend with proper upgrade handling
		if strings.HasPrefix(r.URL.Path, "/ws") {
			proxyWebSocket(w, r, c.Host, c.Port)
			return
		}

		// Serve static files with SPA fallback
		if err == nil {
			app.SPAHandler(spaFS).ServeHTTP(w, r)
		} else {
			http.Error(w, "SPA not available", http.StatusServiceUnavailable)
		}
	})

	// Layer middlewares: Gzip → CacheControl → Handler
	httpsHandler := middleware.Gzip(middleware.CacheControl(baseHTTPSHandler))

	// Start HTTP server on port 80
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

	// Start HTTPS server on port 443
	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: httpsHandler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
			NextProtos:     []string{"h2", "http/1.1", "acme-tls/1"},
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

	// Graceful shutdown handling
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

// proxyWebSocket handles WebSocket upgrade and bidirectional proxying
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
