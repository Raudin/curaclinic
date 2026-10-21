package handlers

import (
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
)

func mapAppointment(item *domain.Appointment) dto.AppointmentResponse {
	return dto.AppointmentResponse{
		ID: item.ID.String(),
		Doctor: dto.AppointmentDoctor{
			ID:        item.Doctor.ID.String(),
			FirstName: item.Doctor.FirstName,
			LastName:  item.Doctor.LastName,
			Specialty: item.Doctor.Specialty,
		},
		Patient: dto.PersonResponse{
			ID:        item.Patient.ID.String(),
			FirstName: item.Patient.FirstName,
			LastName:  item.Patient.LastName,
			Email:     item.Patient.Email,
		},
		StartTime:   item.StartTime,
		EndTime:     item.EndTime,
		Status:      item.Status,
		Reason:      item.Reason,
		CancelledAt: item.CancelledAt,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapAppointmentList(items []domain.Appointment) []dto.AppointmentResponse {
	out := make([]dto.AppointmentResponse, 0, len(items))
	for i := range items {
		out = append(out, mapAppointment(&items[i]))
	}
	return out
}

func mapEvents(items []domain.AppointmentEvent) []dto.AppointmentEventResponse {
	out := make([]dto.AppointmentEventResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.AppointmentEventResponse{
			ID:            item.ID,
			EventType:     item.EventType,
			FromStartTime: item.FromStartTime,
			ToStartTime:   item.ToStartTime,
			Reason:        item.EventReason,
			ActorName:     item.ActorName,
			CreatedAt:     item.CreatedAt,
		})
	}
	return out
}

func mapRoles(items []domain.Role) []dto.RoleResponse {
	out := make([]dto.RoleResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.RoleResponse{
			ID:          item.ID.String(),
			Scope:       item.Scope,
			Name:        item.Name,
			Description: item.Description,
			IsSystem:    item.IsSystem,
			Permissions: emptyIfNil(item.Permissions),
			CreatedAt:   item.CreatedAt,
		})
	}
	return out
}

func mapPermissions(items []domain.Permission) []dto.PermissionResponse {
	out := make([]dto.PermissionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.PermissionResponse{
			ID:          item.ID.String(),
			Scope:       item.Scope,
			Resource:    item.Resource,
			Action:      item.Action,
			Code:        item.Code,
			Description: item.Description,
		})
	}
	return out
}

func mapOverrides(items []domain.PermissionOverride) []dto.OverrideResponse {
	out := make([]dto.OverrideResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapOverride(item))
	}
	return out
}

func mapOverride(item domain.PermissionOverride) dto.OverrideResponse {
	return dto.OverrideResponse{
		ID:           item.ID.String(),
		PermissionID: item.PermissionID.String(),
		Permission:   item.PermissionCode,
		Effect:       item.Effect,
		Note:         item.Note,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func mapAuditLogs(items []domain.AuditLog) []dto.AuditLogResponse {
	out := make([]dto.AuditLogResponse, 0, len(items))
	for _, item := range items {
		actorID := ""
		if item.ActorID != nil {
			actorID = item.ActorID.String()
		}
		metadata := item.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		out = append(out, dto.AuditLogResponse{
			ID:           item.ID,
			ActorID:      actorID,
			Action:       item.Action,
			ResourceType: item.ResourceType,
			ResourceID:   item.ResourceID,
			Metadata:     metadata,
			RequestID:    item.RequestID,
			CreatedAt:    item.CreatedAt,
		})
	}
	return out
}
