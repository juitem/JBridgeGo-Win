package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	path string
	mu   sync.Mutex
}

func NewStorage() (*Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(home, ".jbridgego-win")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		_ = os.MkdirAll(appDir, 0755)
	}
	return &Storage{path: filepath.Join(appDir, "settings.json")}, nil
}

func (s *Storage) Load() (*AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return NewDefaultState(), nil
	}
	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return NewDefaultState(), nil
	}
	if state.UrlAliases == nil { state.UrlAliases = make(map[string]string) }
	if state.HostZoomLevels == nil { state.HostZoomLevels = make(map[string]int) }
	return &state, nil
}

func (s *Storage) Save(state *AppState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
