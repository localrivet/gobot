package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"gobot/internal/db"
	"gobot/internal/mcp/mcpauth"
	"gobot/internal/mcp/mcpctx"
	"gobot/internal/svc"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler handles MCP HTTP requests with Bearer token authentication.
type Handler struct {
	svc             *svc.ServiceContext
	authenticator   *mcpauth.Authenticator
	httpHandler     http.Handler
	resourceMetaURL string

	// sessionCache stores MCP servers + ToolContext by session ID (in-memory).
	// Each session has its own server with its own ToolContext (org selection).
	// This allows multiple clients per user with different orgs.
	sessionCache sync.Map // map[sessionID]*sessionData

	// orgSelectionCache is an in-memory cache of org selections.
	// Backed by mcp_sessions DB table for persistence across restarts.
	orgSelectionCache sync.Map // map[sessionID]string (org ID)
}

// sessionData holds cached session data.
type sessionData struct {
	server  *mcp.Server
	toolCtx *mcpctx.ToolContext
}

// NewHandler creates a new MCP handler with authentication.
// baseURL is used to construct the resource metadata URL for OAuth discovery.
func NewHandler(svc *svc.ServiceContext, baseURL string) *Handler {
	baseURL = strings.TrimSuffix(baseURL, "/")

	h := &Handler{
		svc:             svc,
		authenticator:   mcpauth.NewAuthenticator(svc),
		resourceMetaURL: baseURL + "/.well-known/oauth-protected-resource",
	}

	// Create the streamable HTTP handler in STATELESS mode.
	// Stateless mode means the SDK doesn't validate session IDs - we handle it ourselves.
	// This allows sessions to survive server restarts because:
	// 1. Client continues using their old session ID after restart
	// 2. SDK doesn't reject it (stateless)
	// 3. We restore org selection from our persistent orgSelectionStore
	streamHandler := mcp.NewStreamableHTTPHandler(
		h.getServerForRequest,
		&mcp.StreamableHTTPOptions{
			Stateless: true, // Don't track sessions in SDK - we do it ourselves
		},
	)

	// Wrap with our custom auth middleware that properly formats WWW-Authenticate
	h.httpHandler = h.authMiddleware(streamHandler)

	return h
}

// authMiddleware validates Bearer tokens and returns proper OAuth challenge on 401.
// This custom implementation ensures WWW-Authenticate header is RFC 9728 compliant
// with quoted resource_metadata value.
// It also ensures session IDs are generated and communicated back to the client.
func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for session ID - generate one if not provided
		sessionID := r.Header.Get("Mcp-Session-Id")
		newSession := false
		if sessionID == "" {
			sessionID = uuid.New().String()
			newSession = true
			// Add session ID to request so getServerForRequest can use it
			r.Header.Set("Mcp-Session-Id", sessionID)
		}

		fmt.Printf("[MCP DEBUG] %s %s | Session: %q (new: %v) | Accept: %s\n",
			r.Method, r.URL.Path, sessionID, newSession, r.Header.Get("Accept"))

		// Extract Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			h.writeUnauthorized(w, "missing bearer token")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			h.writeUnauthorized(w, "empty bearer token")
			return
		}

		// Verify token using our authenticator (TokenVerifier returns a func)
		tokenInfo, err := h.authenticator.TokenVerifier()(r.Context(), token, r)
		if err != nil {
			h.writeUnauthorized(w, "invalid token")
			return
		}

		// Always set session ID in response header so client can use it for subsequent requests
		// This is critical for MCP session continuity - the client needs to know the session ID
		w.Header().Set("Mcp-Session-Id", sessionID)

		// Add token info to request context and continue
		ctx := mcpauth.ContextWithTokenInfo(r.Context(), tokenInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeUnauthorized sends a 401 response with WWW-Authenticate header for OAuth discovery.
// Uses simple format matching working MCP OAuth implementations.
func (h *Handler) writeUnauthorized(w http.ResponseWriter, msg string) {
	// Bearer challenge with resource_metadata URL and scope
	wwwAuth := fmt.Sprintf(`Bearer resource_metadata="%s", scope="mcp:full"`, h.resourceMetaURL)
	w.Header().Set("WWW-Authenticate", wwwAuth)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// getServerForRequest returns a cached server for the session, or creates a new one.
// We cache by SESSION ID to support multiple clients per user with different orgs.
// In stateless mode, the SDK doesn't track sessions - we do it ourselves.
// The authMiddleware ensures a session ID is always present in the request header.
// This allows org selection to survive server restarts.
func (h *Handler) getServerForRequest(r *http.Request) *mcp.Server {
	// Session ID is always set by authMiddleware (generated if not provided by client)
	sessionID := r.Header.Get("Mcp-Session-Id")

	// Get token info from context (set by authMiddleware)
	tokenInfo := mcpauth.TokenInfoFromContext(r.Context())
	if tokenInfo == nil {
		fmt.Println("[MCP] ERROR: No token info in context for getServerForRequest")
		return NewServer(h.svc, r)
	}

	// Check cache first - if we have this session, reuse it
	if cached, ok := h.sessionCache.Load(sessionID); ok {
		data := cached.(*sessionData)
		fmt.Printf("[MCP] Using cached session: %s (hasOrg: %v)\n", sessionID, data.toolCtx != nil && data.toolCtx.HasOrg())
		return data.server
	}

	// Session not in cache - this is either a new session or after server restart
	fmt.Printf("[MCP] Session %s not in cache - creating new server\n", sessionID)

	// Create callback that persists org selection for this session
	onOrgSelect := func(userID, orgID string) {
		h.StoreOrgSelection(sessionID, userID, orgID)
	}
	server, toolCtx := NewServerWithContext(h.svc, r, onOrgSelect)

	// Try to restore org selection - 3-level fallback:
	// 1. Memory cache (fastest)
	// 2. DB by session ID (after restart)
	// 3. User's most recent org selection (if session ID changed)
	var orgIDToRestore string
	var found bool

	if storedOrgID, ok := h.orgSelectionCache.Load(sessionID); ok {
		orgIDToRestore = storedOrgID.(string)
		found = true
		fmt.Printf("[MCP] Found org in memory cache for session %s\n", sessionID)
	} else {
		// Not in memory - try DB by session ID (this happens after server restart)
		dbSession, err := h.svc.DB.GetMCPSession(r.Context(), sessionID)
		if err == nil && dbSession.OrgID.Valid {
			orgIDToRestore = dbSession.OrgID.String
			found = true
			// Cache it for next time
			h.orgSelectionCache.Store(sessionID, orgIDToRestore)
			fmt.Printf("[MCP] Found org in DB for session %s\n", sessionID)
		} else if toolCtx != nil && toolCtx.UserID() != "" {
			// 3rd level fallback: Get user's most recent org selection
			// This handles the case where client's session ID changed
			userSession, err := h.svc.DB.GetMCPSessionByUser(r.Context(), toolCtx.UserID())
			if err == nil && userSession.OrgID.Valid {
				orgIDToRestore = userSession.OrgID.String
				found = true
				// Cache under new session ID
				h.orgSelectionCache.Store(sessionID, orgIDToRestore)
				// Persist under new session ID (async)
				go func() {
					_ = h.svc.DB.UpsertMCPSession(context.Background(), db.UpsertMCPSessionParams{
						SessionID: sessionID,
						UserID:    toolCtx.UserID(),
						OrgID:     userSession.OrgID,
					})
				}()
				fmt.Printf("[MCP] Found org from user's recent session for %s\n", sessionID)
			}
		}
	}

	if found && toolCtx != nil {
		// Fetch the org from DB and restore selection
		org, err := h.svc.DB.GetOrganizationByID(r.Context(), orgIDToRestore)
		if err == nil {
			// Set the org directly without callback (it's already persisted)
			toolCtx.RestoreOrg(org)
			fmt.Printf("[MCP] Restored org selection for session %s: org=%s\n", sessionID, org.Name)
		} else {
			fmt.Printf("[MCP] Failed to restore org %s for session %s: %v\n", orgIDToRestore, sessionID, err)
		}
	}

	// Cache the session
	h.sessionCache.Store(sessionID, &sessionData{server: server, toolCtx: toolCtx})

	return server
}

// StoreOrgSelection persists org selection for a session.
// Called by org.select tool to ensure selection survives server restarts.
// Stores in both in-memory cache and DB for persistence.
func (h *Handler) StoreOrgSelection(sessionID string, userID, orgID string) {
	// Store in memory cache
	h.orgSelectionCache.Store(sessionID, orgID)

	// Persist to DB (async to not block tool execution)
	go func() {
		ctx := context.Background()
		err := h.svc.DB.UpsertMCPSession(ctx, db.UpsertMCPSessionParams{
			SessionID: sessionID,
			UserID:    userID,
			OrgID:     sql.NullString{String: orgID, Valid: orgID != ""},
		})
		if err != nil {
			fmt.Printf("[MCP] Failed to persist session to DB: %v\n", err)
		} else {
			fmt.Printf("[MCP] Persisted org selection to DB: session=%s org=%s\n", sessionID, orgID)
		}
	}()
}

// ClearOrgSelection removes org selection for a session.
func (h *Handler) ClearOrgSelection(sessionID string) {
	h.orgSelectionCache.Delete(sessionID)

	// Remove from DB (async)
	go func() {
		ctx := context.Background()
		_ = h.svc.DB.DeleteMCPSession(ctx, sessionID)
	}()
}

// ServeHTTP handles all MCP HTTP requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.httpHandler.ServeHTTP(w, r)
}

// Authenticator returns the authenticator for cache invalidation.
func (h *Handler) Authenticator() *mcpauth.Authenticator {
	return h.authenticator
}
