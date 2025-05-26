package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sanek1/GophKeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(login, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        uuid.New(),
		Login:     login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `INSERT INTO users (id, login, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = r.db.Exec(query, user.ID, user.Login, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) GetByLogin(login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE login = $1`
	err := r.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepositoryImpl) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
