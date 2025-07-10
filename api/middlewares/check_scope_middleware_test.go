package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/container-adm-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupRouterWithMiddleware(mw gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	return r
}

func TestCheckScopeMiddleware_WithValidScope(t *testing.T) {
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("claims", &utils.Claims{Scopes: []string{"admin"}})
		c.Next()
	})
	router.Use(CheckScopeMiddleware("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "ok")
}

func TestCheckScopeMiddleware_MissingClaims(t *testing.T) {
	router := setupRouterWithMiddleware(CheckScopeMiddleware("admin"))

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "No scopes found in claims")
}

func TestCheckScopeMiddleware_InvalidClaimsFormat(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", "not-a-claims-object") // wrong type
		c.Next()
	})
	router.Use(CheckScopeMiddleware("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "Invalid claims format")
}

func TestCheckScopeMiddleware_InsufficientScope(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", &utils.Claims{Scopes: []string{"user"}}) // doesn't include "admin"
		c.Next()
	})
	router.Use(CheckScopeMiddleware("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "insufficient scope")
}
