package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	jwtpkg "github.com/yourorg/auth-service/pkg/jwt"
	"github.com/yourorg/auth-service/internal/model"
	"github.com/yourorg/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrEmailTaken         = errors.New("email already registered")
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwtpkg.Manager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, req.Email, string(hash), req.Name, "user")
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	return s.buildAuthResponse(user)
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildAuthResponse(user)
}

// RefreshTokens issues a new access token from a valid refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, jwtpkg.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return s.buildAuthResponse(user)
}

func (s *AuthService) buildAuthResponse(user *model.User) (*model.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.AccessTokenExpiry(),
		User: &model.UserInfo{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}, nil
}
