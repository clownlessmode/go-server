package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	accessdomain "project/internal/modules/access/domain"
	userdomain "project/internal/modules/user/domain"
	sharedauth "project/internal/shared/auth"
)

const (
	CurrentUserIDKey    = "currentUserID"
	CurrentUserLoginKey = "currentUserLogin"
	CurrentUserRoleKey  = "currentUserRole"
)

func Auth(tokenManager *sharedauth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := tokenManager.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		c.Set(CurrentUserIDKey, claims.UserID)
		c.Set(CurrentUserLoginKey, claims.Login)
		c.Set(CurrentUserRoleKey, userdomain.Role(claims.Role))
		c.Next()
	}
}

func RequireRole(role userdomain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole, ok := c.Get(CurrentUserRoleKey)
		if !ok || currentRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

func RequireBankAccess(accessRepo accessdomain.Repository, bankID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentRole, _ := c.Get(CurrentUserRoleKey)
		if currentRole == userdomain.RoleAdmin {
			c.Next()
			return
		}

		userIDValue, ok := c.Get(CurrentUserIDKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing current user"})
			return
		}

		userID, ok := userIDValue.(int64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid current user"})
			return
		}

		hasAccess, err := accessRepo.HasActiveAccess(c.Request.Context(), userID, bankID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "bank access required"})
			return
		}

		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
