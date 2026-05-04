package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var wechatHTTPClient = &http.Client{Timeout: 10 * time.Second}

func WeChatFreePublishBatchGetHandler(w http.ResponseWriter, r *http.Request) {
	NewWeChatFreePublishBatchGetHandler(wechatHTTPClient, "http://api.weixin.qq.com")(w, r)
}

func NewWeChatFreePublishBatchGetHandler(client *http.Client, upstreamBase string) http.HandlerFunc {
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
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
