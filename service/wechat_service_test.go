package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFreePublishBatchGetHandlerCallsOpenAPI(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/cgi-bin/freepublish/batchget" {
			t.Fatalf("path = %s, want /cgi-bin/freepublish/batchget", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %s, want application/json", got)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(body)) != `{"offset":0,"count":20,"no_content":1}` {
			t.Fatalf("body = %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"total_count":0,"item_count":0,"item":[]}`))
	}))
	defer upstream.Close()

	handler := NewWeChatFreePublishBatchGetHandler(upstream.Client(), upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/wechat/freepublish/batchget", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("response content-type = %s, want application/json", got)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"total_count":0,"item_count":0,"item":[]}` {
		t.Fatalf("response body = %s", rec.Body.String())
	}
}

func TestFreePublishBatchGetHandlerUsesTokenFromCloudRunOpenAPI(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			if r.Method != http.MethodGet {
				t.Fatalf("token method = %s, want GET", r.Method)
			}
			if got := r.URL.Query().Get("grant_type"); got != "client_credential" {
				t.Fatalf("grant_type = %s, want client_credential", got)
			}
			if got := r.URL.Query().Get("appid"); got != "app-id" {
				t.Fatalf("appid = %s, want app-id", got)
			}
			if got := r.URL.Query().Get("secret"); got != "app-secret" {
				t.Fatalf("secret = %s, want app-secret", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("x-openapi-seqid", "seq-token")
			w.Write([]byte(`{"access_token":"token-123","expires_in":7200}`))
		case "/cgi-bin/freepublish/batchget":
			if r.Method != http.MethodPost {
				t.Fatalf("batchget method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("access_token"); got != "token-123" {
				t.Fatalf("access_token = %s, want token-123", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"total_count":0,"item_count":0,"item":[]}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer upstream.Close()

	handler := NewWeChatFreePublishBatchGetHandler(upstream.Client(), upstream.URL, WeChatCredentials{
		AppID:     "app-id",
		AppSecret: "app-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/wechat/freepublish/batchget", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"total_count":0,"item_count":0,"item":[]}` {
		t.Fatalf("response body = %s", rec.Body.String())
	}
}

func TestWeChatTokenCheckHandlerDoesNotExposeToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/token" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-openapi-seqid", "seq-token")
		w.Write([]byte(`{"access_token":"token-123","expires_in":7200}`))
	}))
	defer upstream.Close()

	handler := NewWeChatTokenCheckHandler(upstream.Client(), upstream.URL, WeChatCredentials{
		AppID:     "app-id",
		AppSecret: "app-secret",
	})
	req := httptest.NewRequest(http.MethodGet, "/wechat/token/check", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "token-123") {
		t.Fatalf("response body exposes token: %s", rec.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["token_ok"] != true {
		t.Fatalf("token_ok = %v, want true", body["token_ok"])
	}
	if body["openapi_seqid"] != "seq-token" {
		t.Fatalf("openapi_seqid = %v, want seq-token", body["openapi_seqid"])
	}
}

func TestWeChatConfigCheckHandlerReportsPresenceWithoutSecrets(t *testing.T) {
	t.Setenv("WECHAT_APP_ID", "app-id")
	t.Setenv("WECHAT_APP_SECRET", "app-secret")

	req := httptest.NewRequest(http.MethodGet, "/wechat/config/check", nil)
	rec := httptest.NewRecorder()

	WeChatConfigCheckHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.Contains(rec.Body.String(), "app-secret") {
		t.Fatalf("response body exposes secret: %s", rec.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["wechat_app_id_present"] != true {
		t.Fatalf("wechat_app_id_present = %v, want true", body["wechat_app_id_present"])
	}
	if body["wechat_app_secret_present"] != true {
		t.Fatalf("wechat_app_secret_present = %v, want true", body["wechat_app_secret_present"])
	}
}
