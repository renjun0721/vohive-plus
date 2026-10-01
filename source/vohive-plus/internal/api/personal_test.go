package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yibaiba/hideck/internal/config"
)

func TestPersonalUninstallCannotDeleteAuthenticatedUserData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{auth: config.WebConfig{Username: "admin", Password: "fixture-secret"}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/system/uninstall", nil)
	context.Request.Header.Set("Authorization", "Bearer "+testSessionToken(t, "fixture-secret", time.Now().Add(time.Hour)))
	server.handleUninstall(context)
	if recorder.Code != http.StatusGone {
		t.Fatalf("status = %d, want disabled", recorder.Code)
	}
}

func TestPersonalDiagnosticsRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{auth: config.WebConfig{Username: "admin", Password: "fixture-secret"}}
	recorder := httptest.NewRecorder()
	server.newRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/personal/health", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}
