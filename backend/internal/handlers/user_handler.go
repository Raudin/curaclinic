package handlers

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
	"curaclinic/backend/internal/service"
)

// MyProfile returns the current user's profile.
func (h *Handlers) MyProfile(c *gin.Context) {
	actor := actorOf(c)
	user, err := h.Users.Get(c.Request.Context(), actor.UserID)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUser(*user))
}

// UpdateMyProfile updates the current user's names and phone.
func (h *Handlers) UpdateMyProfile(c *gin.Context) {
	var req dto.UpdateUserProfileRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	user, err := h.Users.UpdateProfile(c.Request.Context(), actor, actor.UserID, service.UpdateUserProfileInput{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		DateOfBirth: req.DateOfBirth,
		AvatarURL:   req.AvatarURL,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUserRecord(*user))
}

// UpdateUserProfile lets an authorized admin update any shared user identity.
func (h *Handlers) UpdateUserProfile(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.UpdateUserProfileRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	user, err := h.Users.UpdateProfile(c.Request.Context(), actorOf(c), id, service.UpdateUserProfileInput{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		DateOfBirth: req.DateOfBirth,
		AvatarURL:   req.AvatarURL,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUserRecord(*user))
}

// ListUsers returns a page of clinic users.
func (h *Handlers) ListUsers(c *gin.Context) {
	limit, offset := parsePagination(c)
	role := c.Query("role")
	if role != "" && role != "admin" && role != "doctor" && role != "patient" {
		httpx.BadRequest(c, h.Log, "request validation failed", map[string]any{"role": "must be admin, doctor or patient"})
		return
	}
	items, total, err := h.Users.List(c.Request.Context(), c.Query("search"), role, limit, offset)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	out := make([]dto.UserResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapUser(item))
	}
	httpx.Page(c, out, httpx.Pagination{Limit: limit, Offset: offset, Total: total})
}

// GetUser returns one user with roles.
func (h *Handlers) GetUser(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	user, err := h.Users.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUser(*user))
}

// CreateUser creates a clinic user with roles.
func (h *Handlers) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	roleIDs, ok := parseUUIDList(c, h.Log, "role_ids", req.RoleIDs)
	if !ok {
		return
	}
	actor := actorOf(c)
	user, err := h.Users.Create(c.Request.Context(), actor, service.CreateUserInput{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		DateOfBirth: req.DateOfBirth,
		AvatarURL:   req.AvatarURL,
		Status:      req.Status,
		RoleIDs:     roleIDs,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapUser(*user))
}

// UpdateUserStatus activates or deactivates an account.
func (h *Handlers) UpdateUserStatus(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.UpdateUserStatusRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	user, err := h.Users.SetStatus(c.Request.Context(), actor, id, req.Status, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUserRecord(*user))
}

// AssignUserRoles replaces a user's roles.
func (h *Handlers) AssignUserRoles(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.AssignRolesRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	roleIDs, ok := parseUUIDList(c, h.Log, "role_ids", req.RoleIDs)
	if !ok {
		return
	}
	actor := actorOf(c)
	if err := h.RBAC.AssignRoles(c.Request.Context(), actor.UserID, id, roleIDs, requestctx.Meta(c)); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	user, err := h.Users.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapUser(*user))
}
