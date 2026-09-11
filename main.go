package main

import (
	"log"
	"net/http"
)

const (
	_fileroot = "."
)

func headrerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(_fileroot))))
	mux.HandleFunc("/healthz", headrerReadiness)

	srv := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	log.Printf("Seving on http://localhost%s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
}
