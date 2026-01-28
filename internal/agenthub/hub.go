package agenthub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Frame represents a message frame between SaaS and Agent
type Frame struct {
	Type    string `json:"type"`              // req, res, event
	ID      string `json:"id,omitempty"`      // Request/response correlation ID
	Method  string `json:"method,omitempty"`  // For requests
	Params  any    `json:"params,omitempty"`  // Request parameters
	OK      bool   `json:"ok,omitempty"`      // Response success
	Payload any    `json:"payload,omitempty"` // Response data
	Error   string `json:"error,omitempty"`   // Error message
}

// AgentConnection represents a connected agent
type AgentConnection struct {
	ID        string
	OrgID     string
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	CreatedAt time.Time

	mu sync.Mutex
}

// Hub manages agent connections
type Hub struct {
	// Registered agents by org ID
	agents sync.Map // map[orgID]map[agentID]*AgentConnection

	// Register channel
	register chan *AgentConnection

	// Unregister channel
	unregister chan *AgentConnection

	upgrader websocket.Upgrader
}

// NewHub creates a new agent hub
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *AgentConnection),
		unregister: make(chan *AgentConnection),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for agents
			},
		},
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case agent := <-h.register:
			h.addAgent(agent)
		case agent := <-h.unregister:
			h.removeAgent(agent)
		}
	}
}

// addAgent adds an agent to the hub
func (h *Hub) addAgent(agent *AgentConnection) {
	// Get or create org map
	orgAgentsI, _ := h.agents.LoadOrStore(agent.OrgID, &sync.Map{})
	orgAgents := orgAgentsI.(*sync.Map)
	orgAgents.Store(agent.ID, agent)

	fmt.Printf("[AgentHub] Agent connected: %s (org: %s)\n", agent.ID, agent.OrgID)
}

// removeAgent removes an agent from the hub
func (h *Hub) removeAgent(agent *AgentConnection) {
	if orgAgentsI, ok := h.agents.Load(agent.OrgID); ok {
		orgAgents := orgAgentsI.(*sync.Map)
		orgAgents.Delete(agent.ID)
	}

	close(agent.Send)
	if agent.Conn != nil {
		agent.Conn.Close()
	}

	fmt.Printf("[AgentHub] Agent disconnected: %s (org: %s)\n", agent.ID, agent.OrgID)
}

// GetAgent returns an agent by org and agent ID
func (h *Hub) GetAgent(orgID, agentID string) *AgentConnection {
	if orgAgentsI, ok := h.agents.Load(orgID); ok {
		orgAgents := orgAgentsI.(*sync.Map)
		if agentI, ok := orgAgents.Load(agentID); ok {
			return agentI.(*AgentConnection)
		}
	}
	return nil
}

// GetAgentsForOrg returns all agents for an organization
func (h *Hub) GetAgentsForOrg(orgID string) []*AgentConnection {
	var agents []*AgentConnection
	if orgAgentsI, ok := h.agents.Load(orgID); ok {
		orgAgents := orgAgentsI.(*sync.Map)
		orgAgents.Range(func(_, value any) bool {
			agents = append(agents, value.(*AgentConnection))
			return true
		})
	}
	return agents
}

// SendToAgent sends a frame to a specific agent
func (h *Hub) SendToAgent(orgID, agentID string, frame *Frame) error {
	agent := h.GetAgent(orgID, agentID)
	if agent == nil {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}

	select {
	case agent.Send <- data:
		return nil
	default:
		return fmt.Errorf("agent send buffer full")
	}
}

// BroadcastToOrg sends a frame to all agents in an organization
func (h *Hub) BroadcastToOrg(orgID string, frame *Frame) {
	data, err := json.Marshal(frame)
	if err != nil {
		return
	}

	agents := h.GetAgentsForOrg(orgID)
	for _, agent := range agents {
		select {
		case agent.Send <- data:
		default:
			// Skip if buffer full
		}
	}
}

// HandleWebSocket handles a WebSocket connection from an agent
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, orgID, agentID, userID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[AgentHub] Upgrade error: %v\n", err)
		return
	}

	agent := &AgentConnection{
		ID:        agentID,
		OrgID:     orgID,
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		CreatedAt: time.Now(),
	}

	h.register <- agent

	// Start reader and writer goroutines
	go h.readPump(agent)
	go h.writePump(agent)
}

// readPump reads messages from the agent
func (h *Hub) readPump(agent *AgentConnection) {
	defer func() {
		h.unregister <- agent
	}()

	agent.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	agent.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	agent.Conn.SetPongHandler(func(string) error {
		agent.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := agent.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("[AgentHub] Read error: %v\n", err)
			}
			break
		}

		// Parse frame
		var frame Frame
		if err := json.Unmarshal(message, &frame); err != nil {
			fmt.Printf("[AgentHub] Invalid frame: %v\n", err)
			continue
		}

		// Handle frame
		h.handleFrame(agent, &frame)
	}
}

// writePump writes messages to the agent
func (h *Hub) writePump(agent *AgentConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		agent.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-agent.Send:
			agent.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				agent.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := agent.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			agent.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := agent.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleFrame processes an incoming frame from an agent
func (h *Hub) handleFrame(agent *AgentConnection, frame *Frame) {
	fmt.Printf("[AgentHub] Frame from %s: type=%s method=%s\n", agent.ID, frame.Type, frame.Method)

	switch frame.Type {
	case "res":
		// Response to a request we sent - could be handled via callbacks
	case "event":
		// Event from agent - could be broadcast to other systems
	case "req":
		// Request from agent - handle and respond
		h.handleRequest(agent, frame)
	}
}

// handleRequest handles a request from an agent
func (h *Hub) handleRequest(agent *AgentConnection, frame *Frame) {
	var response Frame
	response.Type = "res"
	response.ID = frame.ID

	switch frame.Method {
	case "ping":
		response.OK = true
		response.Payload = map[string]any{"pong": true, "time": time.Now().Unix()}

	case "status":
		response.OK = true
		response.Payload = map[string]any{
			"agent_id":   agent.ID,
			"org_id":     agent.OrgID,
			"connected":  true,
			"uptime_sec": int(time.Since(agent.CreatedAt).Seconds()),
		}

	default:
		response.OK = false
		response.Error = fmt.Sprintf("unknown method: %s", frame.Method)
	}

	data, _ := json.Marshal(response)
	agent.Send <- data
}
