package handler

import (
	"context"
	"log"
	"sync"

	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
)

// TaskHandler handles task routing and execution
type TaskHandler struct {
	agents map[string]*AgentSession
	mu     sync.RWMutex
}

// AgentSession represents a connected agent
type AgentSession struct {
	ID         string
	LastBeacon int64
	Platform   string
	Hostname   string
	Tasks      chan *protocol.Task
}

// NewTaskHandler creates a new task handler
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		agents: make(map[string]*AgentSession),
	}
}

// RegisterAgent registers a new agent
func (h *TaskHandler) RegisterAgent(session *AgentSession) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.agents[session.ID] = session
	log.Printf("Agent registered: %s", session.ID)
}

// UnregisterAgent removes an agent
func (h *TaskHandler) UnregisterAgent(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.agents, agentID)
	log.Printf("Agent unregistered: %s", agentID)
}

// SendTask sends a task to an agent
func (h *TaskHandler) SendTask(ctx context.Context, task *protocol.Task) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	session, exists := h.agents[task.AgentId]
	if !exists {
		return ErrAgentNotFound
	}

	select {
	case session.Tasks <- task:
		return nil
	default:
		return ErrTaskQueueFull
	}
}

// GetAgent returns an agent session
func (h *TaskHandler) GetAgent(agentID string) (*AgentSession, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, exists := h.agents[agentID]
	return session, exists
}

// ListAgents returns all connected agents
func (h *TaskHandler) ListAgents() []*AgentSession {
	h.mu.RLock()
	defer h.mu.RUnlock()

	agents := make([]*AgentSession, 0, len(h.agents))
	for _, session := range h.agents {
		agents = append(agents, session)
	}
	return agents
}

// Errors
var (
	ErrAgentNotFound  = &HandlerError{Message: "agent not found"}
	ErrTaskQueueFull  = &HandlerError{Message: "task queue full"}
)

// HandlerError represents a handler error
type HandlerError struct {
	Message string
}

func (e *HandlerError) Error() string {
	return e.Message
}