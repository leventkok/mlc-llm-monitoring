package storage

import (
	"errors"
	"sync"

	"github.com/leventkok/mlc-llm-monitoring/internal/models"
)

var (
	ErrUserNotFound      = errors.New("User is not found")
	ErrUsernameTaken     = errors.New("This username is already taken")
	ErrNotFound = errors.New("not found")
)

type MemoryStore struct {
	mu sync.RWMutex

	usersByID       map[string]models.User 
	usernameToID    map[string]string     
}




func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		usersByID:    make(map[string]models.User),
		usernameToID: make(map[string]string),
	}
}

func (s *MemoryStore) Create(user models.User) error {
	s.mu.Lock()         
	defer s.mu.Unlock() 
	
	if _, exist := s.usernameToID[user.Username]; exist {
		return ErrUsernameTaken
	}

	s.usersByID[user.ID] = user
	s.usernameToID[user.Username] = user.ID
	return nil 
}

func (s *MemoryStore) FindByUsername(username string) (models.User, error) {
	s.mu.RLock()        
	defer s.mu.RUnlock()

	
	id, varMi := s.usernameToID[username]
	if !varMi {
		return models.User{}, ErrUserNotFound 
	}
	
	return s.usersByID[id], nil
}

func (s *MemoryStore) FindByID(id string) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, varMi := s.usersByID[id]
	if !varMi {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *MemoryStore) Update(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	eski, ok := s.usersByID[user.ID]
	if !ok {
		return ErrUserNotFound
	}

	if eski.Username != user.Username {
		delete(s.usernameToID, eski.Username)
		s.usernameToID[user.Username] = user.ID
	}

	s.usersByID[user.ID] = user
	return nil
}