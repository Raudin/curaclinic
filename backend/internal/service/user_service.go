package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
	"curaclinic/backend/internal/platform/password"
)

// CreateUserInput creates any clinic user with roles.
type CreateUserInput struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	Phone       string
	DateOfBirth string
	AvatarURL   string
	Status      string
	RoleIDs     []uuid.UUID
}

// UpdateUserProfileInput updates names and phone.
type UpdateUserProfileInput struct {
	FirstName   string
	LastName    string
	Phone       string
	DateOfBirth string
	AvatarURL   string
}

// UserService handles user administration and self-service profile updates.
type UserService struct {
	store      UserStore
	bcryptCost int
}

// NewUserService creates the user service.
func NewUserService(store UserStore, bcryptCost int) *UserService {
	return &UserService{store: store, bcryptCost: bcryptCost}
}

// List returns a page of users.
func (s *UserService) List(ctx context.Context, search, role string, limit, offset int) ([]domain.UserSummary, int64, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.store.ListUsers(ctx, strings.TrimSpace(search), strings.TrimSpace(role), limit, offset)
}

// Get returns a user with roles.
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error) {
	return s.store.GetUserSummary(ctx, id)
}

// Create inserts a user with one or more roles.
func (s *UserService) Create(ctx context.Context, actor *domain.Actor, input CreateUserInput, meta domain.RequestMeta) (*domain.UserSummary, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	firstName, lastName, err := normalizeNames(input.FirstName, input.LastName)
	if err != nil {
		return nil, err
	}
	roleIDs := uniqueIDs(input.RoleIDs)
	dateOfBirth, err := normalizeDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}
	for _, roleID := range roleIDs {
		if _, roleErr := s.store.GetRoleByID(ctx, roleID); roleErr != nil {
			if domain.IsCode(roleErr, domain.CodeNotFound) {
				return nil, domain.ValidationField("role_ids", "one of the selected roles does not exist")
			}
			return nil, roleErr
		}
	}
	if validateErr := password.Validate(input.Password); validateErr != nil {
		return nil, domain.ValidationField("password", validateErr.Error())
	}
	hash, err := password.Hash(input.Password, s.bcryptCost)
	if err != nil {
		return nil, err
	}
	status := input.Status
	if status == "" {
		status = domain.UserStatusActive
	}
	if status != domain.UserStatusActive && status != domain.UserStatusInactive {
		return nil, domain.ValidationField("status", "must be active or inactive")
	}

	return s.store.CreateUserWithRoles(ctx, domain.NewUser{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        strings.TrimSpace(input.Phone),
		DateOfBirth:  dateOfBirth,
		AvatarURL:    strings.TrimSpace(input.AvatarURL),
		Status:       status,
	}, roleIDs, actor.UserID, meta)
}

// UpdateProfile updates a user's own profile, or any profile with the right
// permission.
func (s *UserService) UpdateProfile(ctx context.Context, actor *domain.Actor, userID uuid.UUID, input UpdateUserProfileInput, meta domain.RequestMeta) (*domain.User, error) {
	if actor.UserID != userID && !actor.Can(domain.PermProfilesUpdateAny) {
		return nil, domain.ErrForbidden
	}
	firstName, lastName, err := normalizeNames(input.FirstName, input.LastName)
	if err != nil {
		return nil, err
	}
	dateOfBirth, err := normalizeDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}
	user, err := s.store.UpdateUserProfile(ctx, userID, firstName, lastName, strings.TrimSpace(input.Phone), dateOfBirth, strings.TrimSpace(input.AvatarURL))
	if err != nil {
		return nil, err
	}
	_ = s.store.RecordAudit(ctx, domain.AuditEntry{
		ActorID:      &actor.UserID,
		Action:       domain.AuditProfileUpdated,
		ResourceType: domain.ResourceUser,
		ResourceID:   userID.String(),
		Metadata:     map[string]any{"first_name": firstName, "last_name": lastName},
		Meta:         meta,
	})
	return user, nil
}

// SetStatus activates or deactivates a user while protecting the last admin.
func (s *UserService) SetStatus(ctx context.Context, actor *domain.Actor, userID uuid.UUID, status string, meta domain.RequestMeta) (*domain.User, error) {
	if status != domain.UserStatusActive && status != domain.UserStatusInactive {
		return nil, domain.ValidationField("status", "must be active or inactive")
	}
	if userID == actor.UserID && status == domain.UserStatusInactive {
		return nil, domain.Validation("you cannot deactivate your own account", nil)
	}
	holdsAdmin, err := s.store.HoldsAdminRole(ctx, userID)
	if err != nil {
		return nil, err
	}
	if status == domain.UserStatusInactive && holdsAdmin {
		count, err := s.store.ActiveAdminCount(ctx)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, domain.NewError(domain.CodeConflict, "cannot deactivate the last active admin")
		}
	}
	return s.store.UpdateUserStatus(ctx, userID, status, actor.UserID, meta)
}
