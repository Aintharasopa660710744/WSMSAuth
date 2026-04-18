package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/yourorg/auth-service/pkg/jwt"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "email"
	ContextKeyRole   = "role"
)

// AuthMiddleware validates Bearer JWT on incoming requests.
// Use this in auth-service itself (e.g., /me endpoint) or
// copy the JWT package to other services.
func AuthMiddleware(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format, expected: Bearer <token>",
			})
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			if errors.Is(err, jwtpkg.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token has expired",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Inject claims into context for downstream handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole allows only specific roles to proceed
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if _, ok := allowed[role.(string)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}
