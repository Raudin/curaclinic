package postgres

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/repository/postgres/dbgen"
)

// UpdateDoctorProfile updates profile fields.
func (r *Repo) UpdateDoctorProfile(ctx context.Context, input domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error) {
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		if _, err := q.UpdateDoctorProfile(ctx, dbgen.UpdateDoctorProfileParams{
			UserID:              input.UserID,
			Specialty:           input.Specialty,
			Bio:                 optionalString(input.Bio),
			ExperienceYears:     int32FromInt(input.ExperienceYears),
			Languages:           stringSlice(input.Languages),
			Timezone:            input.Timezone,
			SlotMinutes:         int32FromInt(input.SlotMinutes),
			IsAcceptingPatients: input.IsAcceptingPatients,
		}); err != nil {
			return mapNotFound(err)
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditDoctorProfileUpdated,
			ResourceType: domain.ResourceDoctor,
			ResourceID:   input.UserID.String(),
			Metadata:     map[string]any{"specialty": input.Specialty, "slot_minutes": input.SlotMinutes},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return r.GetDoctor(ctx, input.UserID)
}

// CreateDoctorAccount creates a user, assigns the doctor role and creates the
// doctor profile in one transaction.
func (r *Repo) CreateDoctorAccount(ctx context.Context, user domain.NewUser, roleName string, profile domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error) {
	var createdUserID uuid.UUID
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		row, err := q.CreateUser(ctx, dbgen.CreateUserParams{
			ID:           user.ID,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Phone:        optionalString(user.Phone),
			Status:       dbgen.UserStatus(user.Status),
		})
		if err != nil {
			return mapUniqueConflict(err, "a user with this email already exists")
		}
		createdUserID = row.ID

		role, err := q.GetRoleByName(ctx, roleName)
		if err != nil {
			return mapNotFound(err)
		}
		if _, err := q.CreateDoctorProfile(ctx, dbgen.CreateDoctorProfileParams{
			UserID:              row.ID,
			Specialty:           profile.Specialty,
			Bio:                 optionalString(profile.Bio),
			ExperienceYears:     int32FromInt(profile.ExperienceYears),
			Languages:           stringSlice(profile.Languages),
			Timezone:            profile.Timezone,
			SlotMinutes:         int32FromInt(profile.SlotMinutes),
			IsAcceptingPatients: profile.IsAcceptingPatients,
		}); err != nil {
			return mapUniqueConflict(err, "this user already has a doctor record")
		}
		if err := q.SetSharedIdentity(ctx, dbgen.SetSharedIdentityParams{
			ID:          row.ID,
			DateOfBirth: user.DateOfBirth,
			AvatarUrl:   user.AvatarURL,
		}); err != nil {
			return err
		}
		if err := q.AssignDoctorRole(ctx, dbgen.AssignDoctorRoleParams{
			DoctorID:   row.ID,
			RoleID:     role.ID,
			AssignedBy: &actor,
		}); err != nil {
			return err
		}

		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditDoctorProfileCreated,
			ResourceType: domain.ResourceDoctor,
			ResourceID:   row.ID.String(),
			Metadata:     map[string]any{"email": row.Email, "specialty": profile.Specialty},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return r.GetDoctor(ctx, createdUserID)
}

// CreateDoctorProfileForUser adds doctor access to an existing shared user.
func (r *Repo) CreateDoctorProfileForUser(ctx context.Context, profile domain.NewDoctorProfile, actor uuid.UUID, meta domain.RequestMeta) (*domain.Doctor, error) {
	err := r.WithTx(ctx, func(q *dbgen.Queries) error {
		if _, err := q.CreateDoctorProfile(ctx, dbgen.CreateDoctorProfileParams{
			UserID:              profile.UserID,
			Specialty:           profile.Specialty,
			Bio:                 optionalString(profile.Bio),
			ExperienceYears:     int32FromInt(profile.ExperienceYears),
			Languages:           stringSlice(profile.Languages),
			Timezone:            profile.Timezone,
			SlotMinutes:         int32FromInt(profile.SlotMinutes),
			IsAcceptingPatients: profile.IsAcceptingPatients,
		}); err != nil {
			return mapUniqueConflict(err, "this user already has a doctor record")
		}
		role, err := q.GetRoleByName(ctx, "doctor")
		if err != nil {
			return mapNotFound(err)
		}
		if err := q.AssignDoctorRole(ctx, dbgen.AssignDoctorRoleParams{
			DoctorID:   profile.UserID,
			RoleID:     role.ID,
			AssignedBy: &actor,
		}); err != nil {
			return err
		}
		return r.audit(ctx, q, domain.AuditEntry{
			ActorID:      &actor,
			Action:       domain.AuditDoctorProfileCreated,
			ResourceType: domain.ResourceDoctor,
			ResourceID:   profile.UserID.String(),
			Metadata:     map[string]any{"specialty": profile.Specialty, "linked_existing_user": true},
			Meta:         meta,
		})
	})
	if err != nil {
		return nil, err
	}
	return r.GetDoctor(ctx, profile.UserID)
}
