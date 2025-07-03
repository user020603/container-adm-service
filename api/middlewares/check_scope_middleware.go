package middlewares

import (
	"net/http"
	"thanhnt208/container-adm-service/utils"

	"github.com/gin-gonic/gin"
)

func CheckScopeMiddleware(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsRaw, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: No scopes found in claims",
			})
			return
		}

		claims, ok := claimsRaw.(*utils.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Invalid claims format",
			})
			return
		}

		userScopes := make(map[string]bool)
		for _, scope := range claims.Scopes {
			userScopes[scope] = true
		}

		for _, required := range requiredScopes {
			if userScopes[required] {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Access denied: insufficient scope",
		})
	}
}
