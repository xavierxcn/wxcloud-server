package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadyzRequiresCredentialsInStandardMode(t *testing.T) {
	t.Setenv("WECHAT_API_MODE", "standard")
	t.Setenv("WECHAT_APP_ID", "")
	t.Setenv("WECHAT_APP_SECRET", "")

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	ReadyzHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["ready"] != false {
		t.Fatalf("ready = %v, want false", body["ready"])
	}
}

func TestReadyzDoesNotExposeSecrets(t *testing.T) {
	t.Setenv("WECHAT_API_MODE", "standard")
	t.Setenv("WECHAT_APP_ID", "app-id")
	t.Setenv("WECHAT_APP_SECRET", "app-secret")

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	ReadyzHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "app-secret") {
		t.Fatalf("response body exposes secret: %s", rec.Body.String())
	}
}
