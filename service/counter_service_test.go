package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCounterHandlerReturnsJSONErrorWhenDatabaseConfigMissing(t *testing.T) {
	t.Setenv("MYSQL_ADDRESS", "")
	t.Setenv("MYSQL_USERNAME", "")
	t.Setenv("MYSQL_PASSWORD", "")
	t.Setenv("MYSQL_DATABASE", "")

	req := httptest.NewRequest(http.MethodGet, "/api/count", nil)
	rec := httptest.NewRecorder()

	CounterHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %s, want application/json", got)
	}
	if !strings.Contains(rec.Body.String(), "MYSQL_ADDRESS") {
		t.Fatalf("body = %s, want database config error", rec.Body.String())
	}
}
