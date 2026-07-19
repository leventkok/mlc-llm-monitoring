package storage

import (
	"sync"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

type ConfigStore struct {
	mu     sync.RWMutex
	config models.Config
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{
		config: models.Config{
			AppName: "mlc-llm-monitoring",
			Model:   "gemma-3-1b-it",
			Version: "0.1.0",
		},
	}
}

func (s *ConfigStore) Get() models.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *ConfigStore) Update(c models.Config) models.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = c
	return s.config
}