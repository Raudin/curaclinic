package router_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/dto"
	"curaclinic/backend/internal/httpx"
)

func TestBookingFlowEndToEnd(t *testing.T) {
	app := newTestApp(t)

	login := app.request(t, http.MethodPost, "/api/auth/login", "", dto.LoginRequest{
		Email:    app.patientEmail,
		Password: "PatientPass-12345",
	})
	require.Equal(t, http.StatusOK, login.Code, login.Body.String())
	tokens := decodeData[dto.TokenResponse](t, login)
	require.NotEmpty(t, tokens.AccessToken)
	require.Contains(t, tokens.Permissions, domain.PermAppointmentsCreateSelf)
	access := tokens.AccessToken

	me := app.request(t, http.MethodGet, "/api/auth/me", access, nil)
	require.Equal(t, http.StatusOK, me.Code, me.Body.String())

	refresh := app.request(t, http.MethodPost, "/api/auth/refresh", "", dto.RefreshRequest{
		RefreshToken: tokens.RefreshToken,
	})
	require.Equal(t, http.StatusOK, refresh.Code, refresh.Body.String())
	refreshed := decodeData[dto.TokenResponse](t, refresh)
	require.NotEmpty(t, refreshed.AccessToken)
	require.NotEqual(t, tokens.RefreshToken, refreshed.RefreshToken)

	meAgain := app.request(t, http.MethodGet, "/api/auth/me", refreshed.AccessToken, nil)
	require.Equal(t, http.StatusOK, meAgain.Code, meAgain.Body.String())

	doctors := app.request(t, http.MethodGet, "/api/doctors", access, nil)
	require.Equal(t, http.StatusOK, doctors.Code, doctors.Body.String())
	require.NotEmpty(t, decodeData[[]dto.DoctorResponse](t, doctors))

	availabilityPath := "/api/doctors/" + app.doctorID.String() + "/availability?date=" + app.date
	availability := app.request(t, http.MethodGet, availabilityPath, access, nil)
	require.Equal(t, http.StatusOK, availability.Code, availability.Body.String())
	before := decodeData[dto.AvailabilityResponse](t, availability)
	require.NotEmpty(t, before.Slots)

	start := before.Slots[0].StartTime
	holdRecorder := app.request(t, http.MethodPost, "/api/holds", access, dto.CreateHoldRequest{
		DoctorID:  app.doctorID.String(),
		StartTime: &start,
	})
	require.Equal(t, http.StatusCreated, holdRecorder.Code, holdRecorder.Body.String())
	hold := decodeData[dto.HoldResponse](t, holdRecorder)
	require.NotEmpty(t, hold.Token)

	heldAvailability := app.request(t, http.MethodGet, availabilityPath, access, nil)
	require.Equal(t, http.StatusOK, heldAvailability.Code)
	afterHold := decodeData[dto.AvailabilityResponse](t, heldAvailability)
	require.Len(t, afterHold.Slots, len(before.Slots)-1, "the held slot must disappear from availability")

	bookRecorder := app.request(t, http.MethodPost, "/api/appointments", access, dto.BookAppointmentRequest{
		DoctorID:  app.doctorID.String(),
		StartTime: &start,
		Reason:    "annual wellness checkup",
		HoldToken: hold.Token,
	})
	require.Equal(t, http.StatusCreated, bookRecorder.Code, bookRecorder.Body.String())
	appointment := decodeData[dto.AppointmentResponse](t, bookRecorder)
	require.Equal(t, domain.AppointmentStatusConfirmed, appointment.Status)
	require.Equal(t, app.patientID.String(), appointment.Patient.ID)
	require.Equal(t, app.doctorID.String(), appointment.Doctor.ID)

	duplicate := app.request(t, http.MethodPost, "/api/holds", access, dto.CreateHoldRequest{
		DoctorID:  app.doctorID.String(),
		StartTime: &start,
	})
	require.Equal(t, http.StatusConflict, duplicate.Code, duplicate.Body.String())

	listRecorder := app.request(t, http.MethodGet, "/api/appointments?scope=upcoming", access, nil)
	require.Equal(t, http.StatusOK, listRecorder.Code, listRecorder.Body.String())
	var page struct {
		Data       []dto.AppointmentResponse `json:"data"`
		Pagination httpx.Pagination          `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(listRecorder.Body.Bytes(), &page))
	require.Equal(t, int64(1), page.Pagination.Total)
	require.Len(t, page.Data, 1)
	require.Equal(t, appointment.ID, page.Data[0].ID)

	events := app.request(t, http.MethodGet, "/api/appointments/"+appointment.ID+"/events", access, nil)
	require.Equal(t, http.StatusOK, events.Code, events.Body.String())
	history := decodeData[[]dto.AppointmentEventResponse](t, events)
	require.Len(t, history, 1)
	require.Equal(t, domain.EventCreated, history[0].EventType)

	cancel := app.request(t, http.MethodPatch, "/api/appointments/"+appointment.ID+"/cancel", access, dto.CancelAppointmentRequest{
		Reason: "patient developed a fever",
	})
	require.Equal(t, http.StatusOK, cancel.Code, cancel.Body.String())
	cancelled := decodeData[dto.AppointmentResponse](t, cancel)
	require.Equal(t, domain.AppointmentStatusCancelled, cancelled.Status)
	require.NotNil(t, cancelled.CancelledAt)

	releasedAvailability := app.request(t, http.MethodGet, availabilityPath, access, nil)
	require.Equal(t, http.StatusOK, releasedAvailability.Code)
	afterCancel := decodeData[dto.AvailabilityResponse](t, releasedAvailability)
	require.Len(t, afterCancel.Slots, len(before.Slots), "cancelling must free the slot")

	forbidden := app.request(t, http.MethodGet, "/api/users", access, nil)
	require.Equal(t, http.StatusForbidden, forbidden.Code, forbidden.Body.String())
}
