package domain

// OverrideEffect values.
type OverrideEffect string

// Override effect values.
const (
	OverrideAllow OverrideEffect = "allow"
	OverrideDeny  OverrideEffect = "deny"
)

// UserStatus values.
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// Appointment status values.
const (
	AppointmentStatusConfirmed = "confirmed"
	AppointmentStatusCancelled = "cancelled"
)

// Appointment event types.
const (
	EventCreated     = "created"
	EventCancelled   = "cancelled"
	EventRescheduled = "rescheduled"
)
