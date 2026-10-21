package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

func mapDoctorRow(row dbgen.GetDoctorProfileByUserIDRow) domain.Doctor {
	return domain.Doctor{
		ProfileID:           row.ProfileID,
		UserID:              row.UserID,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		Email:               row.Email,
		Specialty:           row.Specialty,
		Bio:                 stringValue(row.Bio),
		ExperienceYears:     int(row.ExperienceYears),
		Languages:           row.Languages,
		AvatarURL:           stringValue(row.AvatarUrl),
		Timezone:            row.Timezone,
		SlotMinutes:         int(row.SlotMinutes),
		IsAcceptingPatients: row.IsAcceptingPatients,
	}
}

func mapListDoctorRow(row dbgen.ListDoctorsRow) domain.Doctor {
	return domain.Doctor{
		ProfileID:           row.ProfileID,
		UserID:              row.UserID,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		Email:               row.Email,
		Specialty:           row.Specialty,
		Bio:                 stringValue(row.Bio),
		ExperienceYears:     int(row.ExperienceYears),
		Languages:           row.Languages,
		AvatarURL:           stringValue(row.AvatarUrl),
		Timezone:            row.Timezone,
		SlotMinutes:         int(row.SlotMinutes),
		IsAcceptingPatients: row.IsAcceptingPatients,
	}
}

// CreateDoctorProfile inserts a doctor profile.
func (r *Repo) CreateDoctorProfile(ctx context.Context, input domain.NewDoctorProfile) (*domain.Doctor, error) {
	row, err := r.q.CreateDoctorProfile(ctx, dbgen.CreateDoctorProfileParams{
		UserID:              input.UserID,
		Specialty:           input.Specialty,
		Bio:                 optionalString(input.Bio),
		ExperienceYears:     int32FromInt(input.ExperienceYears),
		Languages:           stringSlice(input.Languages),
		Timezone:            input.Timezone,
		SlotMinutes:         int32FromInt(input.SlotMinutes),
		IsAcceptingPatients: input.IsAcceptingPatients,
	})
	if err != nil {
		return nil, mapUniqueConflict(err, "this user already has a doctor profile")
	}
	return r.GetDoctor(ctx, row.UserID)
}

// GetDoctor loads a doctor profile joined with the user record.
func (r *Repo) GetDoctor(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error) {
	row, err := r.q.GetDoctorProfileByUserID(ctx, userID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	doctor := mapDoctorRow(row)
	return &doctor, nil
}

// ListDoctors returns the doctor directory.
func (r *Repo) ListDoctors(ctx context.Context, filter domain.DoctorFilter) ([]domain.Doctor, error) {
	rows, err := r.q.ListDoctors(ctx, dbgen.ListDoctorsParams{
		Specialty: filter.Specialty,
		Accepting: filter.Accepting,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Doctor, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapListDoctorRow(row))
	}
	return items, nil
}
