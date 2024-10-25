package event

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirrobot01/lamba/common"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	Events   map[string]Event
	mu       sync.Mutex
	filePath string
}

func NewManager(configDir string) *Manager {
	filePath := filepath.Join(configDir, "db", "events.json")
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil
	}
	manager := &Manager{
		Events:   make(map[string]Event),
		filePath: filePath,
	}
	_ = manager.load()
	return manager
}

func (m *Manager) load() error {
	// Load the functions from the file
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error reading function data: %v\n", err)
		}
		return err
	}

	err = json.Unmarshal(data, &m.Events)
	if err != nil {
		fmt.Printf("Error parsing function data: %v\n", err)
	}
	return err
}

func (m *Manager) saveToFile() error {
	data, err := json.Marshal(m.Events)
	if err != nil {
		return fmt.Errorf("error encoding function data: %v", err)
	}

	err = os.WriteFile(m.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing function data: %v", err)
	}

	return nil
}

func (m *Manager) Add(trigger, fn, runtime, payload string) Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := Event{
		ID:        uuid.New().String(),
		Trigger:   trigger,
		Function:  fn,
		Runtime:   runtime,
		Payload:   common.ParsePayload(payload),
		StartedAt: time.Now(),
		Started:   true,
	}
	m.Events[e.ID] = e
	go func() {
		_ = m.saveToFile()
	}()
	return e
}

func (m *Manager) Get(id string) (Event, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, exists := m.Events[id]
	return e, exists
}

func (m *Manager) Update(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events[e.ID] = e
	_ = m.saveToFile()
}

func (m *Manager) List() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	events := make([]Event, 0, len(m.Events))
	for _, e := range m.Events {
		events = append(events, e)
	}
	return events
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Events, id)
	_ = m.saveToFile()
}

func (m *Manager) MarkCompleted(e Event) {
	e.Completed = true
	e.CompletedAt = time.Now()
	m.Update(e)
}

func (m *Manager) MarkFailed(e Event, err error) {
	e.Failed = true
	e.FailedAt = time.Now()
	e.ErrorStr = err.Error()
	m.Update(e)
}

type Event struct {
	ID       string   `json:"id"`
	Trigger  string   `json:"trigger"`
	Function string   `json:"function"`
	Runtime  string   `json:"runtime"`
	Payload  any      `json:"payload"`
	Result   any      `json:"result"`
	Debug    []string `json:"debug"`

	// Status
	Started   bool   `json:"started"`
	Completed bool   `json:"completed"`
	Failed    bool   `json:"failed"`
	ErrorStr  string `json:"error"`

	// Metadata
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	FailedAt    time.Time `json:"failed_at"`
}

func (e *Event) ToJSON() string {
	if e == nil {
		return ""
	}

	j, err := json.Marshal(e)
	if err != nil {
		return ""
	}

	return strings.Replace(string(j), "'", "\\'", -1)
}

func (e *Event) GetPayload() string {
	return e.Payload.(string)
}

func (e *Event) GetResult() string {
	return e.Result.(string)
}
