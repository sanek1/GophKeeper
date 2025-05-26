package repository

import (
	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
)

// UserRepository defines interface for working with user repository
type UserRepository interface {
	Create(login, password string) (*models.User, error)
	GetByLogin(login string) (*models.User, error)
	GetByID(id uuid.UUID) (*models.User, error)
	ValidatePassword(user *models.User, password string) bool
}

// SecretRepository defines interface for working with secret repository
type SecretRepository interface {
	Create(userID uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error)
	GetByID(id uuid.UUID) (*models.Secret, error)
	GetByUserID(userID uuid.UUID) ([]*models.Secret, error)
	Update(id uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error)
	Delete(id uuid.UUID) error
}
