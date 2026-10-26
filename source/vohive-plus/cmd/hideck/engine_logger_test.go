package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	enginelogger "github.com/iniwex5/vowifi-go/engine/logger"
	hostlogger "github.com/yibaiba/hideck/pkg/logger"
	"go.uber.org/zap"
)

func TestEngineLogsReachHostFileAndSSE(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	hostlogger.Setup(hostlogger.LogConfig{Filename: path})
	configureEngineLogger()
	if err := enginelogger.Init("error", "json"); err != nil {
		t.Fatal(err)
	}
	sub := hostlogger.GlobalBroadcaster.Subscribe()
	defer hostlogger.GlobalBroadcaster.Unsubscribe(sub)
	enginelogger.Info("engine-file-sse-regression", zap.String("device", "test-device"))
	select {
	case entry := <-sub:
		if entry.Message != "[test-device] engine-file-sse-regression" || !strings.Contains(entry.Caller, "engine_logger_test.go") {
			t.Fatalf("unexpected SSE entry: %+v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("engine entry missing from host SSE")
	}
	if err := enginelogger.L().Sync(); err != nil && !strings.Contains(err.Error(), "invalid argument") {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[test-device] engine-file-sse-regression") || !strings.Contains(string(data), "engine_logger_test.go") {
		t.Fatalf("engine entry/caller missing from host file: %s", data)
	}
	enginelogger.Debug("hidden-debug-regression")
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "hidden-debug-regression") {
		t.Fatal("host info level did not filter debug")
	}
}
