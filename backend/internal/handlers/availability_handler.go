package handlers

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
)

// GetAvailability returns bookable slots for a doctor and date.
func (h *Handlers) GetAvailability(c *gin.Context) {
	doctorID, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	date := c.Query("date")
	if date == "" {
		httpx.BadRequest(c, h.Log, "request validation failed", map[string]any{"date": "is required (YYYY-MM-DD)"})
		return
	}
	result, err := h.Availability.Availability(c.Request.Context(), doctorID, date)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	slots := make([]dto.SlotResponse, 0, len(result.Slots))
	for _, slot := range result.Slots {
		slots = append(slots, dto.SlotResponse{StartTime: slot.StartTime, EndTime: slot.EndTime})
	}
	httpx.OK(c, dto.AvailabilityResponse{
		DoctorID:    result.DoctorID.String(),
		Date:        result.Date,
		Timezone:    result.Timezone,
		SlotMinutes: result.SlotMinutes,
		Slots:       slots,
	})
}

// CreateHold reserves a slot for the caller.
func (h *Handlers) CreateHold(c *gin.Context) {
	var req dto.CreateHoldRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	doctorID, ok := parseUUIDValue(c, h.Log, "doctor_id", req.DoctorID)
	if !ok {
		return
	}
	hold, err := h.Holds.Create(c.Request.Context(), actorOf(c), doctorID, *req.StartTime)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, dto.HoldResponse{
		Token:     hold.Token,
		DoctorID:  hold.DoctorID.String(),
		StartTime: hold.StartUTC,
		ExpiresAt: hold.ExpiresAt,
	})
}

// ReleaseHold removes a hold owned by the caller.
func (h *Handlers) ReleaseHold(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		httpx.BadRequest(c, h.Log, "request validation failed", map[string]any{"token": "is required"})
		return
	}
	if err := h.Holds.Release(c.Request.Context(), actorOf(c), token); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}
