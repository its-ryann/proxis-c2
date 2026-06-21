package storage

import (
	"sync"
	"time"

	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
)

// AgentStorage provides persistence for agent state
type AgentStorage struct {
	agents map[string]*StoredAgent
	mu     sync.RWMutex
}

// StoredAgent represents a stored agent
type StoredAgent struct {
	ID         string
	Hostname   string
	Platform   string
	IpAddress  string
	FirstSeen  time.Time
	LastBeacon time.Time
	Online     bool
}

// NewAgentStorage creates a new agent storage
func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		agents: make(map[string]*StoredAgent),
	}
}

// StoreAgent stores an agent
func (s *AgentStorage) StoreAgent(agent *protocol.AgentInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents[agent.AgentId] = &StoredAgent{
		ID:         agent.AgentId,
		Hostname:   agent.Hostname,
		Platform:   agent.Platform,
		IpAddress:  agent.IpAddress,
		FirstSeen:  time.Unix(agent.FirstSeen, 0),
		LastBeacon: time.Unix(agent.LastBeacon, 0),
		Online:     agent.Online,
	}

	return nil
}

// GetAgent retrieves an agent
func (s *AgentStorage) GetAgent(agentID string) (*StoredAgent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.agents[agentID]
	return agent, exists
}

// DeleteAgent removes an agent
func (s *AgentStorage) DeleteAgent(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, agentID)
}

// ListAgents returns all agents
func (s *AgentStorage) ListAgents() []*StoredAgent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*StoredAgent, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents
}

// UpdateBeacon updates the last beacon time for an agent
func (s *AgentStorage) UpdateBeacon(agentID string, timestamp int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent, exists := s.agents[agentID]; exists {
		agent.LastBeacon = time.Unix(timestamp, 0)
		agent.Online = true
	}
}