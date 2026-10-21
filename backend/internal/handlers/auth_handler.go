package handlers

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
)

// Register creates a patient account and signs the patient in.
func (h *Handlers) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	result, err := h.Auth.Register(c.Request.Context(), req.FirstName, req.LastName, req.Email, req.Phone, req.DateOfBirth, req.Password, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapAuthResult(result))
}

// Login authenticates a user and returns a token pair.
func (h *Handlers) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	result, err := h.Auth.Login(c.Request.Context(), req.Email, req.Password, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAuthResult(result))
}

// Refresh rotates a refresh token.
func (h *Handlers) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	result, err := h.Auth.Refresh(c.Request.Context(), req.RefreshToken, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAuthResult(result))
}

// Logout revokes the supplied refresh token.
func (h *Handlers) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	if err := h.Auth.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}

// Me returns the current account with roles and permissions.
func (h *Handlers) Me(c *gin.Context) {
	actor := actorOf(c)
	account, err := h.Auth.Me(c.Request.Context(), actor.UserID)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAccount(account))
}

// ChangePassword updates the current user's password.
func (h *Handlers) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	if err := h.Auth.ChangePassword(c.Request.Context(), actor.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}
