package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appLogger "blog-BE/src/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestGinLoggerLogsPanicRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observed := observer.New(zap.InfoLevel)
	originalLogger := appLogger.Log
	appLogger.Log = zap.New(core)
	defer func() {
		_ = appLogger.Log.Sync()
		appLogger.Log = originalLogger
	}()

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(GinLogger())
	router.Use(gin.Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Request-ID", "req-123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.Code)
	}

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zap.ErrorLevel {
		t.Fatalf("expected error level log, got %s", entry.Level)
	}
	if entry.Message != "HTTP Request" {
		t.Fatalf("expected message %q, got %q", "HTTP Request", entry.Message)
	}

	contextMap := entry.ContextMap()
	if got := contextMap["path"]; got != "/panic" {
		t.Fatalf("expected path /panic, got %#v", got)
	}
	if got := contextMap["method"]; got != http.MethodGet {
		t.Fatalf("expected method GET, got %#v", got)
	}
	if got := contextMap["status"]; got != int64(http.StatusInternalServerError) {
		t.Fatalf("expected status 500, got %#v", got)
	}
	if got := contextMap["request_id"]; got != "req-123" {
		t.Fatalf("expected request_id req-123, got %#v", got)
	}
	if _, ok := contextMap["latency"]; !ok {
		t.Fatal("expected latency field to be present")
	}
	if _, ok := contextMap["errors"]; !ok {
		t.Fatal("expected errors field to be present for panic request")
	}
}
