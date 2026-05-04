package service

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"service": "wxcloud-server",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mode := getWeChatAPIModeFromEnv()
	credentials := getWeChatCredentialsFromEnv()
	ready := true
	checks := map[string]interface{}{
		"wechat_api_mode":           mode,
		"wechat_app_id_present":     strings.TrimSpace(credentials.AppID) != "",
		"wechat_app_secret_present": strings.TrimSpace(credentials.AppSecret) != "",
	}
	if mode == "standard" && !credentials.configured() {
		ready = false
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]interface{}{
		"ready":  ready,
		"checks": checks,
	})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
