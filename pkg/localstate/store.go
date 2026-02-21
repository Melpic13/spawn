package localstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	stateVersion = 1
	maxLogs      = 5000
)

// Store provides atomic load/update operations for local CLI state.
type Store struct {
	path string
	mu   sync.Mutex
}

// DaemonState captures daemon lifecycle state.
type DaemonState struct {
	Running   bool      `json:"running"`
	StartedAt time.Time `json:"startedAt,omitempty"`
	StoppedAt time.Time `json:"stoppedAt,omitempty"`
}

// AgentRecord is a persisted agent record.
type AgentRecord struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	ConfigPath   string    `json:"configPath,omitempty"`
	State        string    `json:"state"`
	Capabilities []string  `json:"capabilities,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	LastError    string    `json:"lastError,omitempty"`
}

// ToolRecord stores tool registration metadata.
type ToolRecord struct {
	Name         string    `json:"name"`
	SchemaPath   string    `json:"schemaPath"`
	RegisteredAt time.Time `json:"registeredAt"`
}

// MeshMessage captures one channel message.
type MeshMessage struct {
	Channel string    `json:"channel"`
	From    string    `json:"from,omitempty"`
	Payload string    `json:"payload"`
	SentAt  time.Time `json:"sentAt"`
}

// LogEntry is a persisted log message.
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Agent   string    `json:"agent,omitempty"`
	Message string    `json:"message"`
}

// TraceStep represents one trace event.
type TraceStep struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// TraceRecord stores a trace and steps.
type TraceRecord struct {
	ID        string      `json:"id"`
	Agent     string      `json:"agent"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Steps     []TraceStep `json:"steps"`
}

// State is persisted CLI runtime state.
type State struct {
	Version               int                          `json:"version"`
	UpdatedAt             time.Time                    `json:"updatedAt"`
	Daemon                DaemonState                  `json:"daemon"`
	Agents                map[string]AgentRecord       `json:"agents"`
	InstalledCapabilities []string                     `json:"installedCapabilities"`
	CapabilityConfig      map[string]map[string]string `json:"capabilityConfig"`
	Tools                 map[string]ToolRecord        `json:"tools"`
	MeshChannels          map[string][]MeshMessage     `json:"meshChannels"`
	Logs                  []LogEntry                   `json:"logs"`
	Traces                map[string]TraceRecord       `json:"traces"`
	Config                map[string]string            `json:"config"`
}

// DefaultPath returns the default state file path.
func DefaultPath() (string, error) {
	if override := os.Getenv("SPAWN_STATE_FILE"); override != "" {
		return override, nil
	}
	base := os.Getenv("SPAWN_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		base = filepath.Join(home, ".spawn")
	}
	return filepath.Join(base, "state.json"), nil
}

// Open opens store at default path.
func Open() (*Store, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return OpenAt(path), nil
}

// OpenAt opens store at explicit path.
func OpenAt(path string) *Store {
	return &Store{path: path}
}

// Path returns state file path.
func (s *Store) Path() string {
	return s.path
}

// Load returns current persisted state.
func (s *Store) Load() (*State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked()
}

// Update performs atomic read-modify-write.
func (s *Store) Update(fn func(st *State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	if err := fn(st); err != nil {
		return err
	}
	st.UpdatedAt = time.Now().UTC()
	st.trim()
	return s.saveUnlocked(st)
}

func (s *Store) loadUnlocked() (*State, error) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			st := defaultState()
			if err := s.saveUnlocked(st); err != nil {
				return nil, err
			}
			return st, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("decode state file: %w", err)
	}
	st.normalize()
	return &st, nil
}

func (s *Store) saveUnlocked(st *State) error {
	st.normalize()
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state file: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write state tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

func defaultState() *State {
	return &State{
		Version:               stateVersion,
		UpdatedAt:             time.Now().UTC(),
		Agents:                map[string]AgentRecord{},
		InstalledCapabilities: []string{},
		CapabilityConfig:      map[string]map[string]string{},
		Tools:                 map[string]ToolRecord{},
		MeshChannels:          map[string][]MeshMessage{},
		Logs:                  []LogEntry{},
		Traces:                map[string]TraceRecord{},
		Config:                map[string]string{},
	}
}

func (s *State) normalize() {
	if s.Version == 0 {
		s.Version = stateVersion
	}
	if s.Agents == nil {
		s.Agents = map[string]AgentRecord{}
	}
	if s.CapabilityConfig == nil {
		s.CapabilityConfig = map[string]map[string]string{}
	}
	if s.Tools == nil {
		s.Tools = map[string]ToolRecord{}
	}
	if s.MeshChannels == nil {
		s.MeshChannels = map[string][]MeshMessage{}
	}
	if s.Traces == nil {
		s.Traces = map[string]TraceRecord{}
	}
	if s.Config == nil {
		s.Config = map[string]string{}
	}
	if s.InstalledCapabilities == nil {
		s.InstalledCapabilities = []string{}
	}
	if s.Logs == nil {
		s.Logs = []LogEntry{}
	}
}

func (s *State) trim() {
	s.normalize()
	if len(s.Logs) > maxLogs {
		s.Logs = append([]LogEntry(nil), s.Logs[len(s.Logs)-maxLogs:]...)
	}
	s.InstalledCapabilities = dedupeStrings(s.InstalledCapabilities)
}

func dedupeStrings(input []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(input))
	for _, item := range input {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
