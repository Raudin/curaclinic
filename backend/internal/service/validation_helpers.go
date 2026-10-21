package service

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"curaclinic/backend/internal/domain"
)

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 254 {
		return "", domain.ValidationField("email", "must be a valid email address")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || address.Name != "" {
		return "", domain.ValidationField("email", "must be a valid email address")
	}
	return value, nil
}

func normalizeNames(firstName, lastName string) (string, string, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	if firstName == "" || len(firstName) > 100 {
		return "", "", domain.ValidationField("first_name", "is required and must be at most 100 characters")
	}
	if lastName == "" || len(lastName) > 100 {
		return "", "", domain.ValidationField("last_name", "is required and must be at most 100 characters")
	}
	return firstName, lastName, nil
}

func normalizeDateOfBirth(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	date, err := time.Parse("2006-01-02", value)
	if err != nil || date.After(time.Now()) {
		return "", domain.ValidationField("date_of_birth", "must be a valid date no later than today")
	}
	return date.Format("2006-01-02"), nil
}

func loadLocation(timezone string) (*time.Location, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	return location, nil
}
