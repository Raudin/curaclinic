// Package dto contains HTTP request and response shapes.
package dto

import "time"

// LoginRequest is the credential payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

// RefreshRequest rotates a refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=512"`
}

// LogoutRequest revokes a refresh token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"max=512"`
}

// ChangePasswordRequest changes the current user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,max=72"`
	NewPassword     string `json:"new_password" binding:"required,min=12,max=72"`
}

// RegisterRequest creates a patient account.
type RegisterRequest struct {
	FirstName   string `json:"first_name" binding:"required,max=100"`
	LastName    string `json:"last_name" binding:"required,max=100"`
	Email       string `json:"email" binding:"required,email,max=254"`
	Phone       string `json:"phone" binding:"max=30"`
	DateOfBirth string `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	Password    string `json:"password" binding:"required,min=12,max=72"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone,omitempty"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	DateOfBirth string    `json:"date_of_birth,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenResponse is returned by login and refresh.
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
	Roles        []string     `json:"roles"`
	Permissions  []string     `json:"permissions"`
}

// AccountResponse is the current-user payload.
type AccountResponse struct {
	User        UserResponse `json:"user"`
	Roles       []string     `json:"roles"`
	Permissions []string     `json:"permissions"`
}
