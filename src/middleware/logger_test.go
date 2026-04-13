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

func setupObservedLogger(t *testing.T) *observer.ObservedLogs {
	t.Helper()

	core, observed := observer.New(zap.InfoLevel)
	originalLogger := appLogger.L()
	appLogger.Set(zap.New(core))
	t.Cleanup(func() {
		_ = appLogger.L().Sync()
		appLogger.Set(originalLogger)
	})

	return observed
}

func newTestLoggerRouter() *gin.Engine {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(GinLogger())
	router.Use(gin.Recovery())
	return router
}

func TestGinLoggerLogsSuccessfulRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	observed := setupObservedLogger(t)

	router := newTestLoggerRouter()
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zap.InfoLevel {
		t.Fatalf("expected info level log, got %s", entry.Level)
	}

	contextMap := entry.ContextMap()
	if got := contextMap["path"]; got != "/ok" {
		t.Fatalf("expected path /ok, got %#v", got)
	}
	if got := contextMap["status"]; got != int64(http.StatusOK) {
		t.Fatalf("expected status 200, got %#v", got)
	}
	requestID, ok := contextMap["request_id"].(string)
	if !ok || requestID == "" {
		t.Fatalf("expected generated request_id, got %#v", contextMap["request_id"])
	}
	if _, ok := contextMap["errors"]; ok {
		t.Fatal("did not expect errors field for successful request")
	}
}

func TestGinLoggerLogsPanicRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	observed := setupObservedLogger(t)

	router := newTestLoggerRouter()
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
