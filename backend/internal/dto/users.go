package dto

import "time"

// CreateUserRequest creates a clinic user.
type CreateUserRequest struct {
	Email       string   `json:"email" binding:"required,email,max=254"`
	Password    string   `json:"password" binding:"required,min=12,max=72"`
	FirstName   string   `json:"first_name" binding:"required,max=100"`
	LastName    string   `json:"last_name" binding:"required,max=100"`
	Phone       string   `json:"phone" binding:"max=30"`
	DateOfBirth string   `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	AvatarURL   string   `json:"avatar_url" binding:"max=500"`
	Status      string   `json:"status" binding:"omitempty,oneof=active inactive"`
	RoleIDs     []string `json:"role_ids" binding:"dive,uuid"`
}

// UpdateUserProfileRequest updates names and phone.
type UpdateUserProfileRequest struct {
	FirstName   string `json:"first_name" binding:"required,max=100"`
	LastName    string `json:"last_name" binding:"required,max=100"`
	Phone       string `json:"phone" binding:"max=30"`
	DateOfBirth string `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	AvatarURL   string `json:"avatar_url" binding:"max=500"`
}

// UpdateUserStatusRequest activates or deactivates an account.
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

// AssignRolesRequest replaces a user's roles.
type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"dive,uuid"`
}

// SetOverrideRequest grants or denies a permission.
type SetOverrideRequest struct {
	PermissionID string `json:"permission_id" binding:"required,uuid"`
	Effect       string `json:"effect" binding:"required,oneof=allow deny"`
	Note         string `json:"note" binding:"max=500"`
}

// CreateRoleRequest creates a role.
type CreateRoleRequest struct {
	Scope         string   `json:"scope" binding:"required,oneof=doctor admin"`
	Name          string   `json:"name" binding:"required,max=100"`
	Description   string   `json:"description" binding:"max=500"`
	PermissionIDs []string `json:"permission_ids" binding:"dive,uuid"`
}

// UpdateRoleRequest updates a role.
type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"required,max=100"`
	Description   string   `json:"description" binding:"max=500"`
	PermissionIDs []string `json:"permission_ids" binding:"dive,uuid"`
}

// RoleResponse is the admin representation of a role.
type RoleResponse struct {
	ID          string    `json:"id"`
	Scope       string    `json:"scope"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionResponse is one permission in the catalog.
type PermissionResponse struct {
	ID          string `json:"id"`
	Scope       string `json:"scope"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// OverrideResponse is a user-specific allow/deny rule.
type OverrideResponse struct {
	ID           string    `json:"id"`
	PermissionID string    `json:"permission_id"`
	Permission   string    `json:"permission"`
	Effect       string    `json:"effect"`
	Note         string    `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AccessResponse combines effective access with explicit overrides.
type AccessResponse struct {
	Roles       []string           `json:"roles"`
	Permissions []string           `json:"permissions"`
	Overrides   []OverrideResponse `json:"overrides"`
}
