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
			AppName: "app-review-monitoring",
			Model:   "gemma-2-2b-it-q4f16_1-MLC",
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