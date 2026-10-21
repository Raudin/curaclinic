package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
	"curaclinic/backend/internal/service"
)

// GetDoctorWorkingHours returns a doctor's weekly schedule.
func (h *Handlers) GetDoctorWorkingHours(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	hours, err := h.Doctors.WorkingHours(c.Request.Context(), actorOf(c), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, gin.H{"intervals": mapWorkingHours(hours)})
}

// ReplaceDoctorWorkingHours replaces a doctor's weekly schedule.
func (h *Handlers) ReplaceDoctorWorkingHours(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	h.replaceWorkingHours(c, id)
}

// MySchedule returns the caller's working hours and upcoming blocks.
func (h *Handlers) MySchedule(c *gin.Context) {
	actor := actorOf(c)
	hours, err := h.Doctors.WorkingHours(c.Request.Context(), actor, actor.UserID)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	now := time.Now().UTC()
	until := now.AddDate(0, 0, h.Config.BookingHorizonDays)
	blocks, err := h.Doctors.ListBlocks(c.Request.Context(), actor, actor.UserID, &now, &until)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, gin.H{
		"working_hours": mapWorkingHours(hours),
		"blocks":        mapBlocks(blocks),
	})
}

// ReplaceMyWorkingHours replaces the caller's weekly schedule.
func (h *Handlers) ReplaceMyWorkingHours(c *gin.Context) {
	h.replaceWorkingHours(c, actorOf(c).UserID)
}

// GetDoctorBlocks lists blocks for a doctor.
func (h *Handlers) GetDoctorBlocks(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	h.listBlocks(c, id)
}

// MyBlocks lists the caller's blocks.
func (h *Handlers) MyBlocks(c *gin.Context) {
	h.listBlocks(c, actorOf(c).UserID)
}

// CreateDoctorBlock creates a block for a doctor.
func (h *Handlers) CreateDoctorBlock(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	h.createBlock(c, id)
}

// CreateMyBlock creates a block for the caller.
func (h *Handlers) CreateMyBlock(c *gin.Context) {
	h.createBlock(c, actorOf(c).UserID)
}

// DeleteBlock deletes a block.
func (h *Handlers) DeleteBlock(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	actor := actorOf(c)
	if err := h.Doctors.DeleteBlock(c.Request.Context(), actor, id, requestctx.Meta(c)); err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	noContent(c)
}

// DeleteMyBlock deletes one of the caller's blocks.
func (h *Handlers) DeleteMyBlock(c *gin.Context) { h.DeleteBlock(c) }

func (h *Handlers) replaceWorkingHours(c *gin.Context, doctorID uuid.UUID) {
	var req dto.WorkingHoursRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	intervals := make([]service.IntervalInput, 0, len(req.Intervals))
	for _, interval := range req.Intervals {
		start, err := parseClock(interval.StartTime)
		if err != nil {
			httpx.WriteError(c, h.Log, err)
			return
		}
		end, err := parseClock(interval.EndTime)
		if err != nil {
			httpx.WriteError(c, h.Log, err)
			return
		}
		intervals = append(intervals, service.IntervalInput{
			Weekday:      interval.Weekday,
			StartMinutes: start,
			EndMinutes:   end,
		})
	}
	actor := actorOf(c)
	hours, err := h.Doctors.ReplaceWorkingHours(c.Request.Context(), actor, doctorID, intervals, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, gin.H{"intervals": mapWorkingHours(hours)})
}

func (h *Handlers) listBlocks(c *gin.Context, doctorID uuid.UUID) {
	from, err := parseOptionalTimeQuery(c, "from")
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	to, err := parseOptionalTimeQuery(c, "to")
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	blocks, err := h.Doctors.ListBlocks(c.Request.Context(), actorOf(c), doctorID, from, to)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapBlocks(blocks))
}

func (h *Handlers) createBlock(c *gin.Context, doctorID uuid.UUID) {
	var req dto.CreateBlockRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	actor := actorOf(c)
	block, err := h.Doctors.CreateBlock(c.Request.Context(), actor, doctorID, *req.StartTime, *req.EndTime, req.Reason, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapBlock(*block))
}
