package memory

import (
	"sync"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
)

type ConfigRepo struct {
	mu     sync.RWMutex
	config model.Config
}

func NewConfigRepo() *ConfigRepo {
	return &ConfigRepo{
		config: model.Config{
			AppName: "app-review-monitoring",
			Model:   "gemma-2-2b-it-q4f16_1-MLC",
			Version: "0.1.0",
		},
	}
}

func (s *ConfigRepo) Get() model.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *ConfigRepo) Update(c model.Config) model.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = c
	return s.config
}
