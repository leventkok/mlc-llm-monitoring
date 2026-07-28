package memory

import (
	"errors"
	"strings"
	"sync"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
)

type ConfigRepo struct {
	mu     sync.RWMutex
	config model.Config
	llm    model.LLMConfig
}

func NewConfigRepo() *ConfigRepo {
	return &ConfigRepo{
		config: model.Config{
			AppName: "app-review-monitoring",
			Model:   "gemma-2-2b-it-q4f16_1-MLC",
			Version: "0.1.0",
		},
		llm: model.DefaultLLMConfig("gemma-2-2b-it-q4f16_1-MLC", ""),
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

func (s *ConfigRepo) GetLLM() model.LLMConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.llm
}

func (s *ConfigRepo) UpdateLLM(c model.LLMConfig) (model.LLMConfig, error) {
	if err := validateLLM(c); err != nil {
		return model.LLMConfig{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ActiveModel = s.llm.ActiveModel
	c.ActiveAdapter = s.llm.ActiveAdapter
	s.llm = c
	return s.llm, nil
}

func (s *ConfigRepo) SetRuntimeLLM(activeModel, activeAdapter string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.llm.ActiveModel = strings.TrimSpace(activeModel)
	s.llm.ActiveAdapter = strings.TrimSpace(activeAdapter)
	if s.config.Model == "" || s.config.Model == "gemma-2-2b-it-q4f16_1-MLC" {
		s.config.Model = s.llm.ActiveModel
	}
}

func validateLLM(c model.LLMConfig) error {
	if strings.TrimSpace(c.SystemPrompt) == "" {
		return errors.New("system_prompt is required")
	}
	if len(c.SystemPrompt) > 8000 {
		return errors.New("system_prompt too long (max 8000)")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	if c.MaxTokens < 1 || c.MaxTokens > 4096 {
		return errors.New("max_tokens must be between 1 and 4096")
	}
	if c.TopP < 0 || c.TopP > 1 {
		return errors.New("top_p must be between 0 and 1")
	}
	return nil
}
