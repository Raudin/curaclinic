// Package handlers contains thin Gin handlers that bind requests, call
// services and map results to HTTP responses.
package handlers

import (
	"fmt"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/service"
)

func mapUser(user domain.UserSummary) dto.UserResponse {
	roles := user.Roles
	if roles == nil {
		roles = []string{}
	}
	return dto.UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		DateOfBirth: user.DateOfBirth,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
		Roles:       roles,
		CreatedAt:   user.CreatedAt,
	}
}

func mapUserRecord(user domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		DateOfBirth: user.DateOfBirth,
		AvatarURL:   user.AvatarURL,
		Status:      user.Status,
		Roles:       []string{},
		CreatedAt:   user.CreatedAt,
	}
}

func mapAccount(account *service.Account) dto.AccountResponse {
	return dto.AccountResponse{
		User:        mapUser(account.User),
		Roles:       emptyIfNil(account.Roles),
		Permissions: emptyIfNil(account.Permissions),
	}
}

func mapAuthResult(result *service.AuthResult) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
		User:         mapUser(result.User),
		Roles:        emptyIfNil(result.Roles),
		Permissions:  emptyIfNil(result.Permissions),
	}
}

func mapDoctor(doctor domain.Doctor) dto.DoctorResponse {
	languages := doctor.Languages
	if languages == nil {
		languages = []string{}
	}
	return dto.DoctorResponse{
		ID:                  doctor.UserID.String(),
		ProfileID:           doctor.ProfileID.String(),
		FirstName:           doctor.FirstName,
		LastName:            doctor.LastName,
		Email:               doctor.Email,
		Specialty:           doctor.Specialty,
		Bio:                 doctor.Bio,
		ExperienceYears:     doctor.ExperienceYears,
		Languages:           languages,
		AvatarURL:           doctor.AvatarURL,
		Timezone:            doctor.Timezone,
		SlotMinutes:         doctor.SlotMinutes,
		IsAcceptingPatients: doctor.IsAcceptingPatients,
	}
}

func mapDoctorList(items []domain.Doctor) []dto.DoctorResponse {
	out := make([]dto.DoctorResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapDoctor(item))
	}
	return out
}

func mapWorkingHour(hour domain.WorkingHour) dto.WorkingHourResponse {
	return dto.WorkingHourResponse{
		ID:        hour.ID.String(),
		Weekday:   hour.Weekday,
		StartTime: minutesToClock(hour.StartMinutes),
		EndTime:   minutesToClock(hour.EndMinutes),
	}
}

func mapWorkingHours(items []domain.WorkingHour) []dto.WorkingHourResponse {
	out := make([]dto.WorkingHourResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapWorkingHour(item))
	}
	return out
}

func mapBlock(block domain.ScheduleBlock) dto.BlockResponse {
	return dto.BlockResponse{
		ID:        block.ID.String(),
		DoctorID:  block.DoctorID.String(),
		StartTime: block.StartTime,
		EndTime:   block.EndTime,
		Reason:    block.Reason,
		CreatedAt: block.CreatedAt,
	}
}

func mapBlocks(items []domain.ScheduleBlock) []dto.BlockResponse {
	out := make([]dto.BlockResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapBlock(item))
	}
	return out
}

func emptyIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func minutesToClock(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}
