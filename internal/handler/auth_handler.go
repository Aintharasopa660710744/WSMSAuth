package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/yourorg/auth-service/pkg/jwt"
	"github.com/yourorg/auth-service/internal/middleware"
	"github.com/yourorg/auth-service/internal/model"
	"github.com/yourorg/auth-service/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwtpkg.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired, please login again"})
			return
		}
		if errors.Is(err, jwtpkg.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /auth/me  (requires AuthMiddleware)
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextKeyUserID)
	email, _ := c.Get(middleware.ContextKeyEmail)
	role, _ := c.Get(middleware.ContextKeyRole)

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

// POST /auth/validate  — used by other microservices to verify a token
func (h *AuthHandler) Validate(c *gin.Context) {
	// AuthMiddleware already validated the token; just return the claims
	h.Me(c)
}
