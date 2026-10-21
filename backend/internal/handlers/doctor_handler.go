package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
	"curaclinic/backend/internal/service"
)

// ListDoctors returns the doctor directory.
func (h *Handlers) ListDoctors(c *gin.Context) {
	filter := domain.DoctorFilter{}
	if specialty := c.Query("specialty"); specialty != "" {
		filter.Specialty = &specialty
	}
	if accepting := c.Query("accepting"); accepting != "" {
		value, err := strconv.ParseBool(accepting)
		if err != nil {
			httpx.BadRequest(c, h.Log, "request validation failed", map[string]any{"accepting": "must be true or false"})
			return
		}
		filter.Accepting = &value
	}
	doctors, err := h.Doctors.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapDoctorList(doctors))
}

// GetDoctor returns one doctor.
func (h *Handlers) GetDoctor(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	doctor, err := h.Doctors.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapDoctor(*doctor))
}

// CreateDoctor creates a doctor account and profile.
func (h *Handlers) CreateDoctor(c *gin.Context) {
	var req dto.CreateDoctorRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	doctor, err := h.Doctors.Create(c.Request.Context(), actor, service.CreateDoctorInput{
		UserID:          req.UserID,
		Email:           req.Email,
		Password:        req.Password,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Phone:           req.Phone,
		DateOfBirth:     req.DateOfBirth,
		Specialty:       req.Specialty,
		Bio:             req.Bio,
		ExperienceYears: req.ExperienceYears,
		Languages:       req.Languages,
		AvatarURL:       req.AvatarURL,
		Timezone:        req.Timezone,
		SlotMinutes:     req.SlotMinutes,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	if req.UserID != "" {
		if err := h.RBAC.Invalidate(c.Request.Context()); err != nil {
			httpx.WriteError(c, h.Log, err)
			return
		}
	}
	httpx.Created(c, mapDoctor(*doctor))
}

// UpdateDoctorProfile updates a doctor profile.
func (h *Handlers) UpdateDoctorProfile(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.UpdateDoctorRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	doctor, err := h.Doctors.UpdateProfile(c.Request.Context(), actor, id, service.UpdateDoctorInput{
		Specialty:           req.Specialty,
		Bio:                 req.Bio,
		ExperienceYears:     req.ExperienceYears,
		Languages:           req.Languages,
		Timezone:            req.Timezone,
		SlotMinutes:         req.SlotMinutes,
		IsAcceptingPatients: req.IsAcceptingPatients,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapDoctor(*doctor))
}
