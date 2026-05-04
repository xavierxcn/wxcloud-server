package main

import (
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
	mux.HandleFunc("/api/count", service.CounterHandler)
	mux.HandleFunc("/wechat/freepublish/batchget", service.WeChatFreePublishBatchGetHandler)
	mux.HandleFunc("/wechat/token/check", service.WeChatTokenCheckHandler)

	return mux
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
