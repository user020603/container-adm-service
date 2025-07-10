package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/container-adm-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var originalParseJWT = utils.ParseJWT

func setupRouterWithJWTMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "authorized"})
	})
	return r
}

func TestJWTAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	router := setupRouterWithJWTMiddleware()

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Missing or invalid token")
}

func TestJWTAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	router := setupRouterWithJWTMiddleware()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token xyz")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Missing or invalid token")
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		return nil, errors.New("invalid token")
	}
	defer func() { utils.ParseJWT = originalParseJWT }()

	router := setupRouterWithJWTMiddleware()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	expectedClaims := &utils.Claims{
		UserID: uint(1),
		Scopes: []string{"read", "write"},
	}
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		return expectedClaims, nil
	}
	defer func() { utils.ParseJWT = originalParseJWT }()

	router := setupRouterWithJWTMiddleware()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer validtoken123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "authorized")
}
