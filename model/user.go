package model

import "time"

// User represents a public user profile returned by the API.
// The password hash is excluded from this model for security.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest defines the expected JSON payload for creating a new user.
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest defines the expected JSON payload for updating a user's
// name or email.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// LoginRequest defines the expected JSON payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
