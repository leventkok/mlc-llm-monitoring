package storage

import "github.com/leventkok/mlc-llm-monitoring/internal/models"

type UserStore interface {
	Create(user models.User) error
	FindByUsername(username  string) (models.User, error)
	
	FindByID(id string)(models.User, error)
	Update(user models.User) error
}