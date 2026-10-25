package main

import (
	"context"
	"net/http"
	"openpak/wfc-common/service"
	"os"
	"time"
	"wwfc/api"
	"wwfc/gamestats"
)

var runtimeHTTP, compatibilityHTTP *http.Server

func startHTTP(bool) {
	mux := http.NewServeMux()
	api.RegisterHandlers(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("/", gamestats.HandleWebRequest)
	addr := os.Getenv("WFC_HTTP_ADDR")
	if addr == "" {
		addr = ":20121"
	}
	runtimeHTTP = &http.Server{Addr: addr, Handler: runtimeHandler(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	upstream, err := service.Proxy(os.Getenv("WFC_NAS_URL"), service.Key())
	if err != nil {
		panic(err)
	}
	compatibilityHTTP = service.Server(":20120", upstream)
	go func() {
		if err := compatibilityHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	go func() {
		if err := runtimeHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}
func stopHTTP() {
	if runtimeHTTP != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		runtimeHTTP.Shutdown(ctx)
		if compatibilityHTTP != nil {
			compatibilityHTTP.Shutdown(ctx)
		}
	}
}

func runtimeHandler(mux http.Handler) http.Handler {
	guard := service.Guard(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Write([]byte("ok\n"))
			return
		}
		guard.ServeHTTP(w, r)
	})
}
