package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yibaiba/hideck/internal/config"
	"github.com/yibaiba/hideck/internal/device"
)

func TestUSSDSessionKeepsOriginalNetwork(t *testing.T) {
	for _, tc := range []struct {
		session      string
		active, want bool
	}{
		{"cs", false, false}, {"cs", true, false}, {" cs ", true, false},
		{"ussd-network-dialog", true, true}, {"ussd-network-dialog", false, true},
		{"", true, true}, {"", false, false},
	} {
		if got := ussdSessionUsesVoWiFi(tc.session, tc.active); got != tc.want {
			t.Fatalf("session=%q active=%t got=%t", tc.session, tc.active, got)
		}
	}
}

func TestIMSUSSDNeverFallsBackToCellularAfterDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, action := range []string{"continue", "cancel"} {
		t.Run(action, func(t *testing.T) {
			pool := device.NewPool(&config.Config{})
			cellular := &ussdProviderBackendStub{}
			setNestedPrivateField(t, pool, []string{"workers"}, map[string]*device.Worker{"test-device": {ID: "test-device", Backend: cellular}})
			server := &Server{pool: pool}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Params = gin.Params{{Key: "device_id", Value: "test-device"}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/devices/test-device/actions/ussd/"+action, strings.NewReader(`{"session_id":"ussd-network-dialog","input":"1"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			if action == "continue" {
				server.handleDeviceMgmtContinueUSSD(ctx)
			} else {
				server.handleDeviceMgmtCancelUSSD(ctx)
			}
			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if cellular.cancelCalled || cellular.continueInput != "" {
				t.Fatal("IMS session was routed to cellular backend")
			}
			if !strings.Contains(recorder.Body.String(), "VoWiFi") {
				t.Fatal(recorder.Body.String())
			}
		})
	}
}
