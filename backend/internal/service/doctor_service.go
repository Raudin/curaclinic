package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
)

// CreateDoctorInput creates a doctor user and profile.
type CreateDoctorInput struct {
	UserID          string
	Email           string
	Password        string
	FirstName       string
	LastName        string
	Phone           string
	DateOfBirth     string
	Specialty       string
	Bio             string
	ExperienceYears int
	Languages       []string
	AvatarURL       string
	Timezone        string
	SlotMinutes     int
}

// UpdateDoctorInput updates a doctor profile.
type UpdateDoctorInput struct {
	Specialty           string
	Bio                 string
	ExperienceYears     int
	Languages           []string
	Timezone            string
	SlotMinutes         int
	IsAcceptingPatients bool
}

// DoctorService handles doctor directory and profile operations.
type DoctorService struct {
	store           DoctorStore
	defaultTimezone string
	bcryptCost      int
}

// NewDoctorService creates the doctor service.
func NewDoctorService(store DoctorStore, defaultTimezone string, bcryptCost int) *DoctorService {
	return &DoctorService{store: store, defaultTimezone: defaultTimezone, bcryptCost: bcryptCost}
}

// List returns the doctor directory.
func (s *DoctorService) List(ctx context.Context, filter domain.DoctorFilter) ([]domain.Doctor, error) {
	return s.store.ListDoctors(ctx, filter)
}

// Get returns one doctor.
func (s *DoctorService) Get(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error) {
	return s.store.GetDoctor(ctx, userID)
}

// Create creates a doctor account with the doctor role.
func (s *DoctorService) Create(ctx context.Context, actor *domain.Actor, input CreateDoctorInput, meta domain.RequestMeta) (*domain.Doctor, error) {
	profile, err := s.buildProfile(input.Specialty, input.Bio, input.ExperienceYears, input.Languages, input.Timezone, input.SlotMinutes)
	if err != nil {
		return nil, err
	}
	if input.UserID != "" {
		userID, parseErr := uuid.Parse(strings.TrimSpace(input.UserID))
		if parseErr != nil {
			return nil, domain.ValidationField("user_id", "must be a valid UUID")
		}
		profile.UserID = userID
		return s.store.CreateDoctorProfileForUser(ctx, *profile, actor.UserID, meta)
	}
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	firstName, lastName, err := normalizeNames(input.FirstName, input.LastName)
	if err != nil {
		return nil, err
	}
	dateOfBirth, err := normalizeDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}
	if validateErr := password.Validate(input.Password); validateErr != nil {
		return nil, domain.ValidationField("password", validateErr.Error())
	}
	hash, err := password.Hash(input.Password, s.bcryptCost)
	if err != nil {
		return nil, err
	}

	profile.UserID = uuid.New()

	return s.store.CreateDoctorAccount(ctx,
		domain.NewUser{
			ID:           profile.UserID,
			Email:        email,
			PasswordHash: hash,
			FirstName:    firstName,
			LastName:     lastName,
			Phone:        strings.TrimSpace(input.Phone),
			DateOfBirth:  dateOfBirth,
			AvatarURL:    strings.TrimSpace(input.AvatarURL),
			Status:       domain.UserStatusActive,
		},
		"doctor",
		*profile,
		actor.UserID,
		meta,
	)
}

// UpdateProfile updates a doctor profile when the actor may edit it.
func (s *DoctorService) UpdateProfile(ctx context.Context, actor *domain.Actor, doctorID uuid.UUID, input UpdateDoctorInput, meta domain.RequestMeta) (*domain.Doctor, error) {
	if !canEditDoctorProfile(actor, doctorID) {
		return nil, domain.ErrForbidden
	}
	profile, err := s.buildProfile(input.Specialty, input.Bio, input.ExperienceYears, input.Languages, input.Timezone, input.SlotMinutes)
	if err != nil {
		return nil, err
	}
	profile.UserID = doctorID
	profile.IsAcceptingPatients = input.IsAcceptingPatients
	return s.store.UpdateDoctorProfile(ctx, *profile, actor.UserID, meta)
}

func (s *DoctorService) buildProfile(specialty, bio string, experience int, languages []string, timezone string, slotMinutes int) (*domain.NewDoctorProfile, error) {
	specialty = strings.TrimSpace(specialty)
	if specialty == "" {
		return nil, domain.ValidationField("specialty", "is required")
	}
	if experience < 0 || experience > 80 {
		return nil, domain.ValidationField("experience_years", "must be between 0 and 80")
	}
	if slotMinutes < 5 || slotMinutes > 180 {
		return nil, domain.ValidationField("slot_minutes", "must be between 5 and 180")
	}
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		timezone = s.defaultTimezone
	}
	if _, err := loadLocation(timezone); err != nil {
		return nil, domain.ValidationField("timezone", "must be a valid IANA timezone")
	}
	cleanedLanguages := make([]string, 0, len(languages))
	for _, language := range languages {
		if trimmed := strings.TrimSpace(language); trimmed != "" {
			cleanedLanguages = append(cleanedLanguages, trimmed)
		}
	}
	if len(cleanedLanguages) > 20 {
		return nil, domain.ValidationField("languages", "must contain at most 20 entries")
	}
	return &domain.NewDoctorProfile{
		Specialty:           specialty,
		Bio:                 strings.TrimSpace(bio),
		ExperienceYears:     experience,
		Languages:           cleanedLanguages,
		Timezone:            timezone,
		SlotMinutes:         slotMinutes,
		IsAcceptingPatients: true,
	}, nil
}

func canEditDoctorProfile(actor *domain.Actor, doctorID uuid.UUID) bool {
	if actor.Can(domain.PermAdminUsersUpdate) {
		return true
	}
	return actor.UserID == doctorID && actor.Can(domain.PermDoctorProfileUpdateOwn)
}
