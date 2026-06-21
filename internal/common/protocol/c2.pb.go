package protocol

import (
	"context"
	"time"
)

// TaskType enum
type TaskType int32

const (
	TaskType_TASK_EXECUTE  TaskType = 0
	TaskType_TASK_UPLOAD   TaskType = 1
	TaskType_TASK_DOWNLOAD TaskType = 2
	TaskType_TASK_KILL     TaskType = 3
)

// AgentHello represents an agent connection request
type AgentHello struct {
	AgentId   string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3"`
	Hostname  string `protobuf:"bytes,2,opt,name=hostname,proto3"`
	Platform  string `protobuf:"bytes,3,opt,name=platform,proto3"`
	IpAddress string `protobuf:"bytes,4,opt,name=ip_address,json=ipAddress,proto3"`
	Timestamp int64  `protobuf:"varint,5,opt,name=timestamp,proto3"`
}

// AgentHelloResponse is the response to an agent hello
type AgentHelloResponse struct {
	Accepted bool   `protobuf:"varint,1,opt,name=accepted,proto3"`
	Message  string `protobuf:"bytes,2,opt,name=message,proto3"`
}

// Beacon represents a periodic agent check-in
type Beacon struct {
	AgentId         string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3"`
	Timestamp       int64  `protobuf:"varint,2,opt,name=timestamp,proto3"`
	EncryptedPayload []byte `protobuf:"bytes,3,opt,name=encrypted_payload,json=encryptedPayload,proto3"`
}

// BeaconResponse is the response to a beacon
type BeaconResponse struct {
	Status  string `protobuf:"bytes,1,opt,name=status,proto3"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3"`
}

// Task represents a command to be executed
type Task struct {
	TaskId        string  `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3"`
	AgentId       string  `protobuf:"bytes,2,opt,name=agent_id,json=agentId,proto3"`
	Type          TaskType `protobuf:"varint,3,opt,name=type,proto3,enum=proxis.c2.v1.TaskType"`
	EncryptedData []byte  `protobuf:"bytes,4,opt,name=encrypted_data,json=encryptedData,proto3"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskId         string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3"`
	AgentId        string `protobuf:"bytes,2,opt,name=agent_id,json=agentId,proto3"`
	Success        bool   `protobuf:"varint,3,opt,name=success,proto3"`
	EncryptedOutput []byte `protobuf:"bytes,4,opt,name=encrypted_output,json=encryptedOutput,proto3"`
	Timestamp      int64  `protobuf:"varint,5,opt,name=timestamp,proto3"`
}

// AgentInfo contains information about a connected agent
type AgentInfo struct {
	AgentId   string `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3"`
	Hostname  string `protobuf:"bytes,2,opt,name=hostname,proto3"`
	Platform  string `protobuf:"bytes,3,opt,name=platform,proto3"`
	IpAddress string `protobuf:"bytes,4,opt,name=ip_address,json=ipAddress,proto3"`
	FirstSeen int64  `protobuf:"varint,5,opt,name=first_seen,json=firstSeen,proto3"`
	LastBeacon int64 `protobuf:"varint,6,opt,name=last_beacon,json=lastBeacon,proto3"`
	Online    bool   `protobuf:"varint,7,opt,name=online,proto3"`
}

// C2Server is the gRPC service interface
type C2Server interface {
	HandleAgentHello(ctx context.Context, req *AgentHello) (*AgentHelloResponse, error)
	HandleBeacon(ctx context.Context, req *Beacon) (*BeaconResponse, error)
}

// UnimplementedC2Server is a stub for the C2Server interface
type UnimplementedC2Server struct{}

// NewAgentHello creates a new AgentHello message
func NewAgentHello(agentID, hostname, platform, ipAddress string) *AgentHello {
	return &AgentHello{
		AgentId:   agentID,
		Hostname:  hostname,
		Platform:  platform,
		IpAddress: ipAddress,
		Timestamp: time.Now().Unix(),
	}
}

// NewBeacon creates a new Beacon message
func NewBeacon(agentID string, payload []byte) *Beacon {
	return &Beacon{
		AgentId:         agentID,
		Timestamp:       time.Now().Unix(),
		EncryptedPayload: payload,
	}
}

// NewTask creates a new Task message
func NewTask(taskID, agentID string, taskType TaskType, data []byte) *Task {
	return &Task{
		TaskId:        taskID,
		AgentId:       agentID,
		Type:          taskType,
		EncryptedData: data,
	}
}

// NewTaskResult creates a new TaskResult message
func NewTaskResult(taskID, agentID string, success bool, output []byte) *TaskResult {
	return &TaskResult{
		TaskId:         taskID,
		AgentId:        agentID,
		Success:        success,
		EncryptedOutput: output,
		Timestamp:      time.Now().Unix(),
	}
}