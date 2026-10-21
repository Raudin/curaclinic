package dto

import "time"

// CreateDoctorRequest creates a doctor user and profile.
type CreateDoctorRequest struct {
	UserID          string   `json:"user_id" binding:"omitempty,uuid"`
	Email           string   `json:"email" binding:"omitempty,email,max=254"`
	Password        string   `json:"password" binding:"omitempty,min=12,max=72"`
	FirstName       string   `json:"first_name" binding:"required,max=100"`
	LastName        string   `json:"last_name" binding:"required,max=100"`
	Phone           string   `json:"phone" binding:"max=30"`
	DateOfBirth     string   `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	Specialty       string   `json:"specialty" binding:"required,max=120"`
	Bio             string   `json:"bio" binding:"max=2000"`
	ExperienceYears int      `json:"experience_years" binding:"gte=0,lte=80"`
	Languages       []string `json:"languages" binding:"max=20,dive,max=60"`
	AvatarURL       string   `json:"avatar_url" binding:"max=500"`
	Timezone        string   `json:"timezone" binding:"max=64"`
	SlotMinutes     int      `json:"slot_minutes" binding:"gte=5,lte=180"`
}

// UpdateDoctorRequest updates a doctor profile.
type UpdateDoctorRequest struct {
	Specialty           string   `json:"specialty" binding:"required,max=120"`
	Bio                 string   `json:"bio" binding:"max=2000"`
	ExperienceYears     int      `json:"experience_years" binding:"gte=0,lte=80"`
	Languages           []string `json:"languages" binding:"max=20,dive,max=60"`
	Timezone            string   `json:"timezone" binding:"max=64"`
	SlotMinutes         int      `json:"slot_minutes" binding:"gte=5,lte=180"`
	IsAcceptingPatients bool     `json:"is_accepting_patients"`
}

// IntervalRequest is one working-hours interval in HH:MM.
type IntervalRequest struct {
	Weekday   int    `json:"weekday" binding:"min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// WorkingHoursRequest replaces the weekly schedule.
type WorkingHoursRequest struct {
	Intervals []IntervalRequest `json:"intervals" binding:"required,dive"`
}

// CreateBlockRequest blocks time.
type CreateBlockRequest struct {
	StartTime *time.Time `json:"start_time" binding:"required"`
	EndTime   *time.Time `json:"end_time" binding:"required"`
	Reason    string     `json:"reason" binding:"required,max=200"`
}

// DoctorResponse is the public doctor representation.
type DoctorResponse struct {
	ID                  string   `json:"id"`
	ProfileID           string   `json:"profile_id"`
	FirstName           string   `json:"first_name"`
	LastName            string   `json:"last_name"`
	Email               string   `json:"email"`
	Specialty           string   `json:"specialty"`
	Bio                 string   `json:"bio,omitempty"`
	ExperienceYears     int      `json:"experience_years"`
	Languages           []string `json:"languages"`
	AvatarURL           string   `json:"avatar_url,omitempty"`
	Timezone            string   `json:"timezone"`
	SlotMinutes         int      `json:"slot_minutes"`
	IsAcceptingPatients bool     `json:"is_accepting_patients"`
}

// WorkingHourResponse is one recurring interval.
type WorkingHourResponse struct {
	ID        string `json:"id"`
	Weekday   int    `json:"weekday"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// BlockResponse is a schedule block.
type BlockResponse struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctor_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// SlotResponse is a bookable slot.
type SlotResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// AvailabilityResponse is returned for a doctor/date query.
type AvailabilityResponse struct {
	DoctorID    string         `json:"doctor_id"`
	Date        string         `json:"date"`
	Timezone    string         `json:"timezone"`
	SlotMinutes int            `json:"slot_minutes"`
	Slots       []SlotResponse `json:"slots"`
}

// CreateHoldRequest reserves a slot.
type CreateHoldRequest struct {
	DoctorID  string     `json:"doctor_id" binding:"required,uuid"`
	StartTime *time.Time `json:"start_time" binding:"required"`
}

// HoldResponse is the hold returned to the client.
type HoldResponse struct {
	Token     string    `json:"token"`
	DoctorID  string    `json:"doctor_id"`
	StartTime time.Time `json:"start_time"`
	ExpiresAt time.Time `json:"expires_at"`
}
