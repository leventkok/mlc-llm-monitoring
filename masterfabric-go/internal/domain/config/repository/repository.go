package repository

import "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"

// ConfigRepository stores in-memory application config.
type ConfigRepository interface {
	Get() model.Config
	Update(cfg model.Config) model.Config
}
