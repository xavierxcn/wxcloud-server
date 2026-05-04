package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerStartsWithoutDatabaseConfig(t *testing.T) {
	t.Setenv("MYSQL_ADDRESS", "")
	t.Setenv("MYSQL_USERNAME", "")
	t.Setenv("MYSQL_PASSWORD", "")
	t.Setenv("MYSQL_DATABASE", "")

	handler := newServer()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestListenAddrUsesPortEnv(t *testing.T) {
	t.Setenv("PORT", "8080")

	if got := listenAddr(); got != ":8080" {
		t.Fatalf("listenAddr() = %s, want :8080", got)
	}
}

func TestListenAddrDefaultsToPort80(t *testing.T) {
	t.Setenv("PORT", "")

	if got := listenAddr(); got != ":80" {
		t.Fatalf("listenAddr() = %s, want :80", got)
	}
}
