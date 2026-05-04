package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"wxcloudrun-golang/service"
)

func main() {
	log.Fatal(http.ListenAndServe(listenAddr(), newServer()))
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", service.IndexHandler)
	mux.HandleFunc("/healthz", service.HealthzHandler)
	mux.HandleFunc("/readyz", service.ReadyzHandler)
	mux.HandleFunc("/api/count", service.CounterHandler)
	mux.HandleFunc("/wechat/freepublish/batchget", service.NewWeChatFreePublishBatchGetHandlerFromEnv())
	mux.HandleFunc("/wechat/config/check", service.WeChatConfigCheckHandler)
	mux.HandleFunc("/wechat/token/check", service.NewWeChatTokenCheckHandlerFromEnv())

	return withAdminToken(mux, os.Getenv("ADMIN_TOKEN"))
}

func listenAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func withAdminToken(next http.Handler, token string) http.Handler {
	if strings.TrimSpace(token) == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldRequireAdminToken(r.URL.Path) && !adminTokenMatches(r.Header.Get("X-Admin-Token"), token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func shouldRequireAdminToken(path string) bool {
	return path == "/readyz" || strings.HasPrefix(path, "/wechat/")
}

func adminTokenMatches(got string, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
