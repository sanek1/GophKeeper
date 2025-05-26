package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gophkeeper/internal/models"
)

type SecretRepositoryImpl struct {
	db *sql.DB
}

func NewSecretRepository(db *sql.DB) SecretRepository {
	return &SecretRepositoryImpl{db: db}
}

func (r *SecretRepositoryImpl) Create(userID uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret := &models.Secret{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      secretType,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `INSERT INTO secrets (id, user_id, type, data, metadata, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, secret.ID, secret.UserID, secret.Type, secret.Data, secret.Metadata, secret.CreatedAt, secret.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (r *SecretRepositoryImpl) GetByID(id uuid.UUID) (*models.Secret, error) {
	secret := &models.Secret{}
	query := `SELECT id, user_id, type, data, metadata, created_at, updated_at FROM secrets WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&secret.ID, &secret.UserID, &secret.Type, &secret.Data, &secret.Metadata, &secret.CreatedAt, &secret.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (r *SecretRepositoryImpl) GetByUserID(userID uuid.UUID) ([]*models.Secret, error) {
	query := `SELECT id, user_id, type, data, metadata, created_at, updated_at FROM secrets WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		secret := &models.Secret{}
		err := rows.Scan(&secret.ID, &secret.UserID, &secret.Type, &secret.Data, &secret.Metadata, &secret.CreatedAt, &secret.UpdatedAt)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

func (r *SecretRepositoryImpl) Update(id uuid.UUID, secretType string, data []byte, metadata string) (*models.Secret, error) {
	secret := &models.Secret{
		ID:        id,
		Type:      secretType,
		Data:      data,
		Metadata:  metadata,
		UpdatedAt: time.Now(),
	}

	query := `UPDATE secrets SET type = $1, data = $2, metadata = $3, updated_at = $4 WHERE id = $5 RETURNING user_id, created_at`
	err := r.db.QueryRow(query, secret.Type, secret.Data, secret.Metadata, secret.UpdatedAt, secret.ID).Scan(&secret.UserID, &secret.CreatedAt)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (r *SecretRepositoryImpl) Delete(id uuid.UUID) error {
	query := `DELETE FROM secrets WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
