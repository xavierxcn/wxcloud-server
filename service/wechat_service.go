package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var wechatHTTPClient = &http.Client{Timeout: 10 * time.Second}

type WeChatCredentials struct {
	AppID     string
	AppSecret string
}

type wechatAccessTokenResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
	SeqID       string `json:"-"`
	RawBody     []byte `json:"-"`
}

func WeChatFreePublishBatchGetHandler(w http.ResponseWriter, r *http.Request) {
	NewWeChatFreePublishBatchGetHandler(wechatHTTPClient, "http://api.weixin.qq.com")(w, r)
}

func WeChatTokenCheckHandler(w http.ResponseWriter, r *http.Request) {
	NewWeChatTokenCheckHandler(wechatHTTPClient, "http://api.weixin.qq.com", getWeChatCredentialsFromEnv())(w, r)
}

func WeChatConfigCheckHandler(w http.ResponseWriter, r *http.Request) {
	credentials := getWeChatCredentialsFromEnv()
	res := map[string]interface{}{
		"wechat_app_id_present":     strings.TrimSpace(credentials.AppID) != "",
		"wechat_app_id_length":      len(credentials.AppID),
		"wechat_app_secret_present": strings.TrimSpace(credentials.AppSecret) != "",
		"wechat_app_secret_length":  len(credentials.AppSecret),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func NewWeChatFreePublishBatchGetHandler(client *http.Client, upstreamBase string, _ ...WeChatCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		body := []byte(`{"offset":0,"count":20,"no_content":1}`)
		upstreamURL := strings.TrimRight(upstreamBase, "/") + "/cgi-bin/freepublish/batchget"

		req, err := http.NewRequest(http.MethodPost, upstreamURL, bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		if seqID := resp.Header.Get("x-openapi-seqid"); seqID != "" {
			w.Header().Set("x-openapi-seqid", seqID)
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func NewWeChatTokenCheckHandler(client *http.Client, upstreamBase string, credentials WeChatCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		if !credentials.configured() {
			http.Error(w, "missing WECHAT_APP_ID or WECHAT_APP_SECRET", http.StatusInternalServerError)
			return
		}

		tokenResult, err := fetchWeChatAccessToken(client, upstreamBase, credentials)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		res := map[string]interface{}{
			"token_ok":       tokenResult.AccessToken != "",
			"expires_in":     tokenResult.ExpiresIn,
			"errcode":        tokenResult.ErrCode,
			"errmsg":         tokenResult.ErrMsg,
			"openapi_seqid":  tokenResult.SeqID,
			"via_cloud_open": tokenResult.SeqID != "",
		}

		w.Header().Set("Content-Type", "application/json")
		if tokenResult.AccessToken == "" {
			w.WriteHeader(http.StatusBadGateway)
		}
		json.NewEncoder(w).Encode(res)
	}
}

func getWeChatCredentialsFromEnv() WeChatCredentials {
	return WeChatCredentials{
		AppID:     os.Getenv("WECHAT_APP_ID"),
		AppSecret: os.Getenv("WECHAT_APP_SECRET"),
	}
}

func (credentials WeChatCredentials) configured() bool {
	return strings.TrimSpace(credentials.AppID) != "" && strings.TrimSpace(credentials.AppSecret) != ""
}

func fetchWeChatAccessToken(client *http.Client, upstreamBase string, credentials WeChatCredentials) (*wechatAccessTokenResult, error) {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(upstreamBase, "/")+"/cgi-bin/token", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("grant_type", "client_credential")
	query.Set("appid", credentials.AppID)
	query.Set("secret", credentials.AppSecret)
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResult wechatAccessTokenResult
	tokenResult.RawBody = body
	tokenResult.SeqID = resp.Header.Get("x-openapi-seqid")
	if err := json.Unmarshal(body, &tokenResult); err != nil {
		return nil, err
	}

	return &tokenResult, nil
}
