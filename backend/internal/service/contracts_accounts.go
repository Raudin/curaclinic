package service

import (
	"context"

	"github.com/google/uuid"

	"curaclinic/backend/internal/domain"
)

// AuthStore is the persistence contract for authentication.
type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetActiveUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetPasswordHash(ctx context.Context, id uuid.UUID) (string, error)
	TouchLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error
	CreateSession(ctx context.Context, input domain.NewSession) (*domain.Session, error)
	GetSessionByHash(ctx context.Context, hash string) (*domain.Session, error)
	RotateSession(ctx context.Context, oldSessionID uuid.UUID, input domain.NewSession) error
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeSessionFamily(ctx context.Context, familyID uuid.UUID) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error
	GetUserSummary(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error)
	EnsurePatientIdentity(ctx context.Context, userID uuid.UUID) error
	GetRoleByName(ctx context.Context, name string) (*domain.Role, error)
	CreateUserWithRoles(ctx context.Context, input domain.NewUser, roleIDs []uuid.UUID, assignedBy uuid.UUID, meta domain.RequestMeta) (*domain.UserSummary, error)
}

// AccessReader resolves cached roles and permissions.
type AccessReader interface {
	Access(ctx context.Context, userID uuid.UUID) (domain.UserAccess, error)
}

// RBACStore is the persistence contract for authorization administration.
type RBACStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	EffectivePermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	RoleNames(ctx context.Context, userID uuid.UUID) ([]string, error)
	UserRoleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListRoles(ctx context.Context) ([]domain.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	ListOverrides(ctx context.Context, userID uuid.UUID) ([]domain.PermissionOverride, error)
	ActiveAdminCount(ctx context.Context) (int64, error)
	HoldsAdminRole(ctx context.Context, userID uuid.UUID) (bool, error)
	CreateRole(ctx context.Context, scope, name, description string, permissionIDs []uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) (*domain.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, name, description string, permissionIDs []uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) (*domain.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error
	AssignUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, assignedBy uuid.UUID, meta domain.RequestMeta) error
	SetOverride(ctx context.Context, userID, permissionID uuid.UUID, effect, note string, grantedBy uuid.UUID, meta domain.RequestMeta) (*domain.PermissionOverride, error)
	DeleteOverride(ctx context.Context, userID, permissionID uuid.UUID, actor uuid.UUID, meta domain.RequestMeta) error
}

// UserStore is the persistence contract for user administration.
type UserStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserSummary(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	RoleNames(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListUsers(ctx context.Context, search, role string, limit, offset int) ([]domain.UserSummary, int64, error)
	CreateUserWithRoles(ctx context.Context, input domain.NewUser, roleIDs []uuid.UUID, assignedBy uuid.UUID, meta domain.RequestMeta) (*domain.UserSummary, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName, phone, dateOfBirth, avatarURL string) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status string, actor uuid.UUID, meta domain.RequestMeta) (*domain.User, error)
	ActiveAdminCount(ctx context.Context) (int64, error)
	HoldsAdminRole(ctx context.Context, userID uuid.UUID) (bool, error)
	RecordAudit(ctx context.Context, entry domain.AuditEntry) error
}
