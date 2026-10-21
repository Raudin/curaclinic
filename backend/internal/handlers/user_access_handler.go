package handlers

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
)

// GetUserAccess returns effective roles, permissions and overrides.
func (h *Handlers) GetUserAccess(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	access, overrides, err := h.RBAC.AccessDetail(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, dto.AccessResponse{
		Roles:       emptyIfNil(access.Roles),
		Permissions: emptyIfNil(access.Permissions),
		Overrides:   mapOverrides(overrides),
	})
}

// SetUserOverride grants or denies one permission.
func (h *Handlers) SetUserOverride(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.SetOverrideRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	permissionID, ok := parseUUIDValue(c, h.Log, "permission_id", req.PermissionID)
	if !ok {
		return
	}
	actor := actorOf(c)
	override, err := h.RBAC.SetOverride(c.Request.Context(), actor.UserID, id, permissionID, req.Effect, req.Note, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapOverride(*override))
}

// DeleteUserOverride removes an explicit grant or deny.
func (h *Handlers) DeleteUserOverride(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	permissionID, ok := parseUUIDParam(c, h.Log, "permissionId")
	if !ok {
		return
	}
	actor := actorOf(c)
	if err := h.RBAC.DeleteOverride(c.Request.Context(), actor.UserID, id, permissionID, requestctx.Meta(c)); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}

// ListRoles returns the role catalog.
func (h *Handlers) ListRoles(c *gin.Context) {
	roles, err := h.RBAC.ListRoles(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapRoles(roles))
}

// CreateRole creates a custom role.
func (h *Handlers) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	permissionIDs, ok := parseUUIDList(c, h.Log, "permission_ids", req.PermissionIDs)
	if !ok {
		return
	}
	actor := actorOf(c)
	role, err := h.RBAC.CreateRole(c.Request.Context(), actor.UserID, req.Scope, req.Name, req.Description, permissionIDs, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapRoles([]domain.Role{*role})[0])
}

// UpdateRole updates a role.
func (h *Handlers) UpdateRole(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.UpdateRoleRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	permissionIDs, ok := parseUUIDList(c, h.Log, "permission_ids", req.PermissionIDs)
	if !ok {
		return
	}
	actor := actorOf(c)
	role, err := h.RBAC.UpdateRole(c.Request.Context(), actor.UserID, id, req.Name, req.Description, permissionIDs, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapRoles([]domain.Role{*role})[0])
}

// DeleteRole removes a non-system role.
func (h *Handlers) DeleteRole(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	actor := actorOf(c)
	if err := h.RBAC.DeleteRole(c.Request.Context(), actor.UserID, id, requestctx.Meta(c)); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}

// ListPermissions returns the permission catalog.
func (h *Handlers) ListPermissions(c *gin.Context) {
	permissions, err := h.RBAC.ListPermissions(c.Request.Context())
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapPermissions(permissions))
}
