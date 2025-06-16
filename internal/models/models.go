package models

import (
	"time"

	"github.com/google/uuid"
)

// User system user
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Secret user secret data
type Secret struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`         // Secret type: password, card, text, file, etc.
	Data      []byte    `json:"data" db:"data"`         // Encrypted data
	Metadata  string    `json:"metadata" db:"metadata"` // Metadata (name, description)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// LoginRequest for user login
type LoginRequest struct {
	Login    string `json:"login" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required,min=8"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Login    string `json:"login" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required,min=8"`
}

// SecretRequest for creating or updating a secret
type SecretRequest struct {
	Type     string `json:"type" example:"password" binding:"required,oneof=password card text file note binary"`
	Data     []byte `json:"data" example:"AQIDBA==" binding:"required"`
	Metadata string `json:"metadata" example:"My bank card" binding:"required"`
}

// PasswordData structure for storing password data
type PasswordData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Website  string `json:"website"`
	Notes    string `json:"notes,omitempty"`
}

// CardData structure for storing card data
type CardData struct {
	CardNumber  string `json:"card_number"`
	CardHolder  string `json:"card_holder"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CVV         string `json:"cvv"`
	BankName    string `json:"bank_name,omitempty"`
	CardType    string `json:"card_type,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// TextData structure for storing text data
type TextData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// FileData structure for storing files
type FileData struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Content     []byte `json:"content"`
}

// NoteData structure for storing notes
type NoteData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags,omitempty"`
}

// SecretTypes defines allowed secret types
var SecretTypes = []string{
	"password", // Passwords
	"card",     // Bank cards
	"text",     // Text data
	"file",     // Files
	"note",     // Notes
	"binary",   // Binary data
}

// ClientInfo contains client information
type ClientInfo struct {
	Version     string `json:"version"`
	BuildDate   string `json:"build_date"`
	CommitHash  string `json:"commit_hash"`
	BuildNumber string `json:"build_number"`
}

// SyncStatus contains synchronization status information
type SyncStatus struct {
	LastSyncTime    time.Time `json:"last_sync_time"`
	ItemsUploaded   int       `json:"items_uploaded"`
	ItemsDownloaded int       `json:"items_downloaded"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}
