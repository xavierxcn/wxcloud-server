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

func TestHealthzDoesNotRequireAdminToken(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	handler := newServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAdminTokenProtectsWeChatRoutes(t *testing.T) {
	t.Setenv("ADMIN_TOKEN", "secret")

	handler := newServer()
	req := httptest.NewRequest(http.MethodGet, "/wechat/config/check", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodGet, "/wechat/config/check", nil)
	req.Header.Set("X-Admin-Token", "secret")
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("authorized status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
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
