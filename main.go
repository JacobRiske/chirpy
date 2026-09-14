package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const (
	_fileroot = "."
)

type apiConfg struct {
	fileSeverHits atomic.Int32
}

func (cfg *apiConfg) hitcountString(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileSeverHits.Load())))
}

func (cfg *apiConfg) resetConfg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileSeverHits.Swap(0)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfg) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileSeverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func headrerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	var (
		apiCfg = &apiConfg{}
		header = http.StripPrefix("/app", http.FileServer(http.Dir(_fileroot)))
	)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(header))
	mux.HandleFunc("GET /metrics", apiCfg.hitcountString)
	mux.HandleFunc("GET /healthz", headrerReadiness)
	mux.Handle("POST /reset", apiCfg.resetConfg(http.HandlerFunc(headrerReadiness)))

	srv := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	log.Printf("Seving on http://localhost%s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
}
