package service

import (
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
