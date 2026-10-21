package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"curaclinic/backend/internal/config"
	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
	"curaclinic/backend/internal/repository/postgres"
)

func seedFlowData(ctx context.Context, t *testing.T, repo *postgres.Repo, cfg *config.Config) (uuid.UUID, uuid.UUID, string, string, time.Time) {
	t.Helper()
	location, err := time.LoadLocation(cfg.ClinicTimezone)
	require.NoError(t, err)
	targetDate := time.Now().In(location).AddDate(0, 0, 7)
	date := targetDate.Format("2006-01-02")
	parsedDate, err := time.ParseInLocation("2006-01-02", date, location)
	require.NoError(t, err)

	hash, err := password.Hash("DoctorPass-12345", cfg.BcryptCost)
	require.NoError(t, err)
	doctorID := uuid.New()
	_, err = repo.CreateDoctorAccount(ctx, domain.NewUser{
		ID:           doctorID,
		Email:        fmt.Sprintf("flow-doctor-%s@curaclinic.test", uuid.NewString()),
		PasswordHash: hash,
		FirstName:    "Flow",
		LastName:     "Doctor",
		Status:       domain.UserStatusActive,
	}, "doctor", domain.NewDoctorProfile{
		UserID:              doctorID,
		Specialty:           "Family Medicine",
		Timezone:            cfg.ClinicTimezone,
		SlotMinutes:         30,
		IsAcceptingPatients: true,
	}, doctorID, domain.RequestMeta{IP: "test"})
	require.NoError(t, err)
	require.NoError(t, repo.ReplaceWorkingHours(ctx, doctorID, []domain.WorkingHour{{
		ID: uuid.New(), DoctorID: doctorID, Weekday: int(parsedDate.Weekday()),
		StartMinutes: 9 * 60, EndMinutes: 12 * 60,
	}}, doctorID, domain.RequestMeta{IP: "test"}))

	patientID := uuid.New()
	patientHash, err := password.Hash("PatientPass-12345", cfg.BcryptCost)
	require.NoError(t, err)
	patientEmail := fmt.Sprintf("flow-patient-%s@curaclinic.test", uuid.NewString())
	_, err = repo.CreateUserWithRoles(ctx, domain.NewUser{
		ID:           patientID,
		Email:        patientEmail,
		PasswordHash: patientHash,
		FirstName:    "Flow",
		LastName:     "Patient",
		Status:       domain.UserStatusActive,
	}, nil, patientID, domain.RequestMeta{IP: "test"})
	require.NoError(t, err)

	slot := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 9, 0, 0, 0, location).UTC()
	return doctorID, patientID, patientEmail, date, slot
}

func (a *testApp) request(t *testing.T, method, path, accessToken string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}
	request := httptest.NewRequest(method, path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	a.engine.ServeHTTP(recorder, request)
	return recorder
}

func decodeData[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var envelope struct {
		Data T `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	return envelope.Data
}
