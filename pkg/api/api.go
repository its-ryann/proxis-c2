package api

import (
	"context"

	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
)

// ServerAPI provides the public API interface for the C2 server
type ServerAPI interface {
	// Agent management
	ListAgents(ctx context.Context) ([]*protocol.AgentInfo, error)
	GetAgent(ctx context.Context, agentID string) (*protocol.AgentInfo, error)
	RemoveAgent(ctx context.Context, agentID string) error

	// Task management
	SendTask(ctx context.Context, task *protocol.Task) error
	GetTaskResult(ctx context.Context, taskID string) (*protocol.TaskResult, error)
	ListTasks(ctx context.Context, agentID string) ([]*protocol.Task, error)
}

// AgentAPI provides the public API interface for the C2 agent
type AgentAPI interface {
	// Connection
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error

	// Communication
	SendBeacon(ctx context.Context, beacon *protocol.Beacon) error
	PollTask(ctx context.Context) (*protocol.Task, error)
	SendResult(ctx context.Context, result *protocol.TaskResult) error
}

// ServerAPIImpl implements the ServerAPI interface
type ServerAPIImpl struct {
	server *C2Server
}

// NewServerAPI creates a new ServerAPI instance
func NewServerAPI(server *C2Server) *ServerAPIImpl {
	return &ServerAPIImpl{server: server}
}

// ListAgents returns all connected agents
func (api *ServerAPIImpl) ListAgents(ctx context.Context) ([]*protocol.AgentInfo, error) {
	// Implementation will be provided in Phase 2
	return nil, nil
}

// GetAgent returns a specific agent
func (api *ServerAPIImpl) GetAgent(ctx context.Context, agentID string) (*protocol.AgentInfo, error) {
	// Implementation will be provided in Phase 2
	return nil, nil
}

// RemoveAgent removes an agent
func (api *ServerAPIImpl) RemoveAgent(ctx context.Context, agentID string) error {
	// Implementation will be provided in Phase 2
	return nil
}

// SendTask sends a task to an agent
func (api *ServerAPIImpl) SendTask(ctx context.Context, task *protocol.Task) error {
	// Implementation will be provided in Phase 2
	return nil
}

// GetTaskResult gets the result of a task
func (api *ServerAPIImpl) GetTaskResult(ctx context.Context, taskID string) (*protocol.TaskResult, error) {
	// Implementation will be provided in Phase 2
	return nil, nil
}

// ListTasks lists all tasks for an agent
func (api *ServerAPIImpl) ListTasks(ctx context.Context, agentID string) ([]*protocol.Task, error) {
	// Implementation will be provided in Phase 2
	return nil, nil
}