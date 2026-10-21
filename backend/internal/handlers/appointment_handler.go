package handlers

import (
	"github.com/gin-gonic/gin"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
	"curaclinic/backend/internal/requestctx"
	"curaclinic/backend/internal/service"
)

// BookAppointment books an appointment for the authenticated patient.
func (h *Handlers) BookAppointment(c *gin.Context) {
	var req dto.BookAppointmentRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	doctorID, ok := parseUUIDValue(c, h.Log, "doctor_id", req.DoctorID)
	if !ok {
		return
	}
	actor := actorOf(c)
	appointment, err := h.Appointments.BookForSelf(c.Request.Context(), actor, service.BookForSelfInput{
		DoctorID:  doctorID,
		StartTime: *req.StartTime,
		Reason:    req.Reason,
		HoldToken: req.HoldToken,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapAppointment(appointment))
}

// BookManualAppointment books on behalf of a patient.
func (h *Handlers) BookManualAppointment(c *gin.Context) {
	var req dto.ManualAppointmentRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	doctorID, ok := parseUUIDValue(c, h.Log, "doctor_id", req.DoctorID)
	if !ok {
		return
	}
	patientID, ok := parseUUIDValue(c, h.Log, "patient_id", req.PatientID)
	if !ok {
		return
	}
	appointment, err := h.Appointments.BookManual(c.Request.Context(), actorOf(c), service.BookManualInput{
		DoctorID:  doctorID,
		PatientID: patientID,
		StartTime: *req.StartTime,
		Reason:    req.Reason,
	}, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Created(c, mapAppointment(appointment))
}

// ListAppointments lists appointments visible to the caller.
func (h *Handlers) ListAppointments(c *gin.Context) {
	doctorID, ok := parseOptionalUUIDQuery(c, h.Log, "doctor_id")
	if !ok {
		return
	}
	patientID, ok := parseOptionalUUIDQuery(c, h.Log, "patient_id")
	if !ok {
		return
	}
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
	limit, offset := parsePagination(c)
	items, total, err := h.Appointments.List(c.Request.Context(), actorOf(c), service.ListAppointmentsInput{
		Scope:     c.Query("scope"),
		As:        c.Query("as"),
		DoctorID:  doctorID,
		PatientID: patientID,
		FromTime:  from,
		ToTime:    to,
		Order:     c.Query("order"),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Page(c, mapAppointmentList(items), httpx.Pagination{Limit: limit, Offset: offset, Total: total})
}

// GetAppointment returns one appointment.
func (h *Handlers) GetAppointment(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	appointment, err := h.Appointments.Get(c.Request.Context(), actorOf(c), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAppointment(appointment))
}

// CancelAppointment cancels an appointment.
func (h *Handlers) CancelAppointment(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.CancelAppointmentRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	appointment, err := h.Appointments.Cancel(c.Request.Context(), actorOf(c), id, req.Reason, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAppointment(appointment))
}

// RescheduleAppointment moves an appointment.
func (h *Handlers) RescheduleAppointment(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	var req dto.RescheduleAppointmentRequest
	if !httpx.BindBody(c, h.Log, &req) {
		return
	}
	appointment, err := h.Appointments.Reschedule(c.Request.Context(), actorOf(c), id, *req.StartTime, req.Reason, requestctx.Meta(c))
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapAppointment(appointment))
}

// AppointmentEvents returns the appointment history.
func (h *Handlers) AppointmentEvents(c *gin.Context) {
	id, ok := parseUUIDParam(c, h.Log, "id")
	if !ok {
		return
	}
	events, err := h.Appointments.Events(c.Request.Context(), actorOf(c), id)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.OK(c, mapEvents(events))
}

// ListAuditLogs returns filtered audit entries.
func (h *Handlers) ListAuditLogs(c *gin.Context) {
	filter := domain.AuditFilter{}
	if value := c.Query("resource_type"); value != "" {
		filter.ResourceType = &value
	}
	if value := c.Query("resource_id"); value != "" {
		filter.ResourceID = &value
	}
	actorID, ok := parseOptionalUUIDQuery(c, h.Log, "actor_id")
	if !ok {
		return
	}
	filter.ActorID = actorID
	limit, offset := parsePagination(c)
	filter.Limit, filter.Offset = limit, offset

	items, total, err := h.Audit.List(c.Request.Context(), filter)
	if err != nil {
		httpx.WriteError(c, h.Log, err)
		return
	}
	httpx.Page(c, mapAuditLogs(items), httpx.Pagination{Limit: limit, Offset: offset, Total: total})
}
