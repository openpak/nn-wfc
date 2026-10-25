package main

import (
	"context"
	"net/http"
	"openpak/wfc-common/service"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	key := service.Key()
	nas, e := service.Proxy(os.Getenv("WFC_NAS_URL"), key)
	if e != nil {
		panic(e)
	}
	api, e := service.Proxy(os.Getenv("WFC_API_URL"), key)
	if e != nil {
		panic(e)
	}
	stats, e := service.Proxy(os.Getenv("WFC_GAMESTATS_URL"), key)
	if e != nil {
		panic(e)
	}
	h := http.NewServeMux()
	h.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	h.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.Host, "gamestats") {
			stats.ServeHTTP(w, r)
			return
		}
		nas.ServeHTTP(w, r)
	}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	addr := os.Getenv("WFC_HTTP_ADDR")
	if addr == "" {
		addr = ":20120"
	}
	if e := service.Run(ctx, service.Server(addr, h)); e != nil {
		panic(e)
	}
}
