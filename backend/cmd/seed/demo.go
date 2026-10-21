package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
	"curaclinic/backend/internal/repository/postgres"
)

type demoDoctor struct {
	email     string
	firstName string
	lastName  string
	specialty string
	bio       string
	languages []string
	avatarURL string
}

func seedDemo(ctx context.Context, repo *postgres.Repo) error {
	demoPassword := envOr("DEMO_PASSWORD", "DemoPass-12345")
	if err := password.Validate(demoPassword); err != nil {
		return err
	}
	cost := envInt("BCRYPT_COST", 12)
	timezone := envOr("CLINIC_TIMEZONE", "Africa/Nairobi")

	doctors := []demoDoctor{
		{"naomi.kamau@curaclinic.local", "Naomi", "Kamau", "Family Medicine", "Preventive care, chronic symptom resolution and metabolic screening.", []string{"English", "Swahili"}, "https://randomuser.me/api/portraits/women/44.jpg"},
		{"daniel.otieno@curaclinic.local", "Daniel", "Otieno", "Internal Medicine", "Adult health and long-term condition management.", []string{"English", "Swahili"}, "https://randomuser.me/api/portraits/men/32.jpg"},
		{"grace.wanjiku@curaclinic.local", "Grace", "Wanjiku", "General Practice", "First-contact care, wellness guidance and coordinated support.", []string{"English", "Kikuyu"}, "https://randomuser.me/api/portraits/women/68.jpg"},
		{"peter.mwangi@curaclinic.local", "Peter", "Mwangi", "Family Medicine", "Routine checkups, ongoing care and personalized prevention plans.", []string{"English", "Swahili"}, "https://randomuser.me/api/portraits/men/75.jpg"},
		{"amina.hassan@curaclinic.local", "Amina", "Hassan", "Internal Medicine", "Careful assessment with clear, evidence-based treatment plans.", []string{"English", "Swahili", "Arabic"}, "https://randomuser.me/api/portraits/women/29.jpg"},
	}

	for _, item := range doctors {
		if _, err := repo.GetUserByEmail(ctx, item.email); err == nil {
			fmt.Printf("doctor %s already exists; skipping\n", item.email)
			continue
		} else if !domain.IsCode(err, domain.CodeNotFound) {
			return err
		}
		hash, err := password.Hash(demoPassword, cost)
		if err != nil {
			return err
		}
		userID := uuid.New()
		doctor, err := repo.CreateDoctorAccount(ctx, domain.NewUser{
			ID:           userID,
			Email:        item.email,
			PasswordHash: hash,
			FirstName:    item.firstName,
			LastName:     item.lastName,
			AvatarURL:    item.avatarURL,
			Status:       domain.UserStatusActive,
		}, "doctor", domain.NewDoctorProfile{
			UserID:              userID,
			Specialty:           item.specialty,
			Bio:                 item.bio,
			ExperienceYears:     8,
			Languages:           item.languages,
			Timezone:            timezone,
			SlotMinutes:         30,
			IsAcceptingPatients: true,
		}, userID, domain.RequestMeta{IP: "seed"})
		if err != nil {
			return err
		}
		if err := repo.ReplaceWorkingHours(ctx, doctor.UserID, weekdayWorkingHours(doctor.UserID), userID, domain.RequestMeta{IP: "seed"}); err != nil {
			return err
		}
		fmt.Printf("created doctor %s\n", item.email)
	}

	patientEmail := "patient@curaclinic.local"
	if _, err := repo.GetUserByEmail(ctx, patientEmail); err == nil {
		fmt.Printf("patient %s already exists; skipping\n", patientEmail)
	} else if !domain.IsCode(err, domain.CodeNotFound) {
		return err
	} else {
		hash, err := password.Hash(demoPassword, cost)
		if err != nil {
			return err
		}
		patientID := uuid.New()
		if _, err := repo.CreateUserWithRoles(ctx, domain.NewUser{
			ID:           patientID,
			Email:        patientEmail,
			PasswordHash: hash,
			FirstName:    "Amelia",
			LastName:     "Hart",
			AvatarURL:    "https://randomuser.me/api/portraits/women/44.jpg",
			Status:       domain.UserStatusActive,
		}, nil, patientID, domain.RequestMeta{IP: "seed"}); err != nil {
			return err
		}
		fmt.Printf("created patient %s\n", patientEmail)
	}

	adminEmail := "admin@curaclinic.local"
	if _, err := repo.GetUserByEmail(ctx, adminEmail); err == nil {
		fmt.Printf("admin %s already exists; skipping\n", adminEmail)
		return nil
	} else if !domain.IsCode(err, domain.CodeNotFound) {
		return err
	}
	role, err := repo.GetRoleByName(ctx, "admin")
	if err != nil {
		return err
	}
	hash, err := password.Hash(demoPassword, cost)
	if err != nil {
		return err
	}
	adminID := uuid.New()
	if _, err := repo.CreateUserWithRoles(ctx, domain.NewUser{
		ID:           adminID,
		Email:        adminEmail,
		PasswordHash: hash,
		FirstName:    "Clinic",
		LastName:     "Admin",
		Status:       domain.UserStatusActive,
	}, []uuid.UUID{role.ID}, adminID, domain.RequestMeta{IP: "seed"}); err != nil {
		return err
	}
	fmt.Printf("created admin %s\n", adminEmail)
	return nil
}

func weekdayWorkingHours(doctorID uuid.UUID) []domain.WorkingHour {
	hours := make([]domain.WorkingHour, 0, 10)
	for weekday := 1; weekday <= 5; weekday++ {
		hours = append(hours,
			domain.WorkingHour{ID: uuid.New(), DoctorID: doctorID, Weekday: weekday, StartMinutes: 9 * 60, EndMinutes: 12 * 60},
			domain.WorkingHour{ID: uuid.New(), DoctorID: doctorID, Weekday: weekday, StartMinutes: 13 * 60, EndMinutes: 16 * 60},
		)
	}
	return hours
}
