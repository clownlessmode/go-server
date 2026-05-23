package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AgentAPIKey(expectedKey string) gin.HandlerFunc {
	expectedKey = strings.TrimSpace(expectedKey)

	return func(c *gin.Context) {
		if expectedKey == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "sms agent api is disabled"})
			return
		}

		key := strings.TrimSpace(c.GetHeader("X-SMS-Agent-Key"))
		if key == "" {
			const prefix = "Bearer "
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, prefix) {
				key = strings.TrimSpace(strings.TrimPrefix(auth, prefix))
			}
		}
		if key != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid sms agent key"})
			return
		}

		c.Next()
	}
}
