package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yibaiba/hideck/internal/config"
	"github.com/yibaiba/hideck/internal/device"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPersonalDeviceConfigAdditionBeyondFiveAllowed(t *testing.T) {
	for _, count := range []int{5, 6, 12, 1000} {
		configs := make([]config.DeviceConfig, count)
		for i := range configs {
			configs[i].ID = fmt.Sprintf("device-%d", i)
		}
		if err := validateFreeDeviceConfigLimit(configs); err != nil {
			t.Fatalf("count %d: %v", count, err)
		}
	}
}

func TestPersonalDeviceListReportsUnlimitedQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{pool: device.NewPool(&config.Config{}), auth: config.WebConfig{Username: "admin", Password: "fixture-secret"}}
	router := gin.New()
	router.GET("/api/devices", server.authMiddleware(), server.handleDeviceMgmtList)
	request := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	request.Header.Set("Authorization", "Bearer "+testSessionToken(t, "fixture-secret", time.Now().Add(time.Hour)))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status %d", response.Code)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	quota, exists := body["device_limit"]
	if !exists || string(quota) != "0" {
		t.Fatalf("unlimited quota = %s", quota)
	}
}
