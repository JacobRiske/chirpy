package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/JacobRiske/chirpy/internal/database"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	_fileroot    = "."
	_defaultAddr = ":8085"
)

type apiConfg struct {
	fileSeverHits atomic.Int32
	db            *database.Queries
	env           string
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

func (cfg *apiConfg) resetConfg(w http.ResponseWriter, r *http.Request) {

	if cfg.env != "dev" {
		log.Printf("env is not in the correct setting")
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText(http.StatusForbidden)))
		return
	}

	cfg.fileSeverHits.Swap(0)

	if err := cfg.restDatabases(context.Background()); err != nil {
		log.Printf("failed to clear a database: %s", err)
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return

	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
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

// POST /api/?
func postApi(s string) string {
	return fmt.Sprintf("POST /api/%s", s)
}

// GET /api/?
func getApi(s string) string {
	return fmt.Sprintf("GET /api/%s", s)
}

func makeDB() (*database.Queries, string, error) {

	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	activeEVN := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, "", err
	}

	return database.New(db), activeEVN, nil
}

func (cfg *apiConfg) restDatabases(ctx context.Context) error {
	if err := cfg.db.ResetUsers(ctx); err != nil {
		return errors.Wrap(err, "failed to reset user database")
	}

	return nil
}
