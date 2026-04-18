package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/auth-service/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name, role string) (*model.User, error) {
	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Email, user.PasswordHash, user.Name,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		// Check unique violation (pgx error code 23505)
		if isPgUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, is_active, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, is_active, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// isPgUniqueViolation checks if the error is a Postgres unique constraint violation
func isPgUniqueViolation(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		containsCode(err.Error(), "23505")
}

func containsCode(s, code string) bool {
	return len(s) >= len(code) && (s == code ||
		(len(s) > len(code) && findSubstr(s, code)))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
