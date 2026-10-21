package dto

import "time"

// BookAppointmentRequest books for the authenticated patient.
type BookAppointmentRequest struct {
	DoctorID  string     `json:"doctor_id" binding:"required,uuid"`
	StartTime *time.Time `json:"start_time" binding:"required"`
	Reason    string     `json:"reason" binding:"required,max=1000"`
	HoldToken string     `json:"hold_token" binding:"required,max=512"`
}

// ManualAppointmentRequest books on behalf of a patient.
type ManualAppointmentRequest struct {
	DoctorID  string     `json:"doctor_id" binding:"required,uuid"`
	PatientID string     `json:"patient_id" binding:"required,uuid"`
	StartTime *time.Time `json:"start_time" binding:"required"`
	Reason    string     `json:"reason" binding:"required,max=1000"`
}

// CancelAppointmentRequest cancels with a reason.
type CancelAppointmentRequest struct {
	Reason string `json:"reason" binding:"required,max=1000"`
}

// RescheduleAppointmentRequest moves an appointment.
type RescheduleAppointmentRequest struct {
	StartTime *time.Time `json:"start_time" binding:"required"`
	Reason    string     `json:"reason" binding:"max=1000"`
}

// PersonResponse is a person summary.
type PersonResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
}

// AppointmentDoctor is a doctor summary inside an appointment.
type AppointmentDoctor struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Specialty string `json:"specialty"`
}

// AppointmentResponse is the public appointment representation.
type AppointmentResponse struct {
	ID          string            `json:"id"`
	Doctor      AppointmentDoctor `json:"doctor"`
	Patient     PersonResponse    `json:"patient"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	Status      string            `json:"status"`
	Reason      string            `json:"reason"`
	CancelledAt *time.Time        `json:"cancelled_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AppointmentEventResponse is one history entry.
type AppointmentEventResponse struct {
	ID            int64      `json:"id"`
	EventType     string     `json:"event_type"`
	FromStartTime *time.Time `json:"from_start_time,omitempty"`
	ToStartTime   *time.Time `json:"to_start_time,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	ActorName     string     `json:"actor_name"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AuditLogResponse is one audit entry.
type AuditLogResponse struct {
	ID           int64          `json:"id"`
	ActorID      string         `json:"actor_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id,omitempty"`
	Metadata     map[string]any `json:"metadata"`
	RequestID    string         `json:"request_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}
