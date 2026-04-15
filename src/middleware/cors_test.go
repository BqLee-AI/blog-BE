package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-BE/src/config"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareUsesLatestConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("APP_ENV", "missing-env")
	t.Setenv("APP_CONFIG_FILE", "")
	t.Setenv("CORS_ALLOW_ORIGINS", "https://initial-client.example.com")

	if err := config.LoadConfig(); err != nil {
		t.Fatalf("initial LoadConfig returned error: %v", err)
	}

	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	t.Setenv("CORS_ALLOW_ORIGINS", "https://web-client.example.com")
	if err := config.LoadConfig(); err != nil {
		t.Fatalf("updated LoadConfig returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Host = "api.example.com"
	req.Header.Set("Origin", "https://web-client.example.com")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "https://web-client.example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin https://web-client.example.com, got %q", got)
	}
}
