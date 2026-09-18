package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	_fileroot    = "."
	_defaultAddr = ":8085"
)

type apiConfg struct {
	fileSeverHits atomic.Int32
}

func (cfg *apiConfg) hitcountString(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("admin.html")
	if err != nil {
		http.Error(w, errors.Wrap(err, "could not load in the page").Error(), http.StatusInternalServerError)
		return
	}

	htmlFormat := fmt.Sprintf(string(data), cfg.fileSeverHits.Load())

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlFormat))
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
		apiCfg    = &apiConfg{}
		appHeader = http.StripPrefix("/app", http.FileServer(http.Dir(_fileroot)))
	)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(appHeader))
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitcountString)
	mux.HandleFunc("GET /api/healthz", headrerReadiness)
	mux.Handle("POST /admin/reset", apiCfg.resetConfg(http.HandlerFunc(headrerReadiness)))

	srv := &http.Server{
		Addr:    _defaultAddr,
		Handler: mux,
	}

	log.Printf("Seving on http://localhost%s\n", _defaultAddr)

	log.Fatal(srv.ListenAndServe())
}
