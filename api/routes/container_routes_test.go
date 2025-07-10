package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"thanhnt208/container-adm-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockHandler implements IContainerHandler
type mockHandler struct {
	Called *bool
}

func (m *mockHandler) handle(c *gin.Context) {
	*m.Called = true
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (m *mockHandler) CreateContainer(c *gin.Context)  { m.handle(c) }
func (m *mockHandler) ViewContainers(c *gin.Context)   { m.handle(c) }
func (m *mockHandler) UpdateContainer(c *gin.Context)  { m.handle(c) }
func (m *mockHandler) DeleteContainer(c *gin.Context)  { m.handle(c) }
func (m *mockHandler) ImportContainers(c *gin.Context) { m.handle(c) }
func (m *mockHandler) ExportContainers(c *gin.Context) { m.handle(c) }

func setupRouterWithScope(scope string, called *bool) *gin.Engine {
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		return &utils.Claims{
			UserID: 1,
			Scopes: []string{scope},
		}, nil
	}

	mock := &mockHandler{Called: called}
	return SetupContainerRoutes(mock)
}

func performRequest(t *testing.T, method, path, scope string) *httptest.ResponseRecorder {
	called := false
	router := setupRouterWithScope(scope, &called)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	router.ServeHTTP(w, req)

	assert.True(t, called, "Handler should be called for "+path)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)

	return w
}

func TestRoutesWithValidScopes(t *testing.T) {
	tests := []struct {
		Method string
		Path   string
		Scope  string
	}{
		{"POST", "/create", "container:create"},
		{"POST", "/view", "container:read"},
		{"PUT", "/update/1", "container:update"},
		{"DELETE", "/delete/1", "container:delete"},
		{"POST", "/import", "container:import"},
		{"POST", "/export", "container:export"},
	}

	for _, tc := range tests {
		t.Run(tc.Path, func(t *testing.T) {
			performRequest(t, tc.Method, tc.Path, tc.Scope)
		})
	}
}

func TestRoutesWithMissingOrInvalidScopes(t *testing.T) {
	tests := []struct {
		Method string
		Path   string
		Scope  string
	}{
		{"POST", "/create", "wrong:scope"},
		{"POST", "/view", "another:scope"},
		{"PUT", "/update/1", "container:read"},
		{"DELETE", "/delete/1", ""},
		{"POST", "/import", "container:read"},
		{"POST", "/export", "container:create"},
	}

	for _, tc := range tests {
		t.Run(tc.Path, func(t *testing.T) {
			called := false
			utils.ParseJWT = func(token string) (*utils.Claims, error) {
				return &utils.Claims{
					UserID: 1,
					Scopes: []string{tc.Scope},
				}, nil
			}

			mock := &mockHandler{Called: &called}
			router := SetupContainerRoutes(mock)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.Method, tc.Path, nil)
			req.Header.Set("Authorization", "Bearer mock-token")
			router.ServeHTTP(w, req)

			assert.False(t, called, "Handler should not be called for "+tc.Path)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "insufficient scope")
		})
	}
}

func TestRoutesWithMissingToken(t *testing.T) {
	router := SetupContainerRoutes(&mockHandler{Called: new(bool)})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/create", nil) // No Authorization header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing or invalid token")
}
