package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JacobRiske/chirpy/apipayload"
	_ "github.com/lib/pq"
)

const (
	_chirps      = "chirps"
	_getonething = "%s/%s"
)

func main() {
	var (
		apiCfg    = &apiConfg{}
		appHeader = http.StripPrefix("/app", http.FileServer(http.Dir(_fileroot)))
		err       error
	)

	if apiCfg.db, apiCfg.env, err = makeDB(); err != nil {
		log.Printf("failed to start the DB: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(appHeader))
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitcountString)
	mux.HandleFunc("GET /api/healthz", headrerReadiness)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetConfg)
	mux.HandleFunc(postApi(_chirps), func(w http.ResponseWriter, r *http.Request) {
		apipayload.CreateChirp(w, r, apiCfg.db)
	})
	mux.HandleFunc(postApi("users"), func(w http.ResponseWriter, r *http.Request) {
		apipayload.AddUser(w, r, apiCfg.db)
	})
	mux.HandleFunc(getApi(_chirps), func(w http.ResponseWriter, r *http.Request) {
		apipayload.GetAllChirps(w, r, apiCfg.db)
	})
	mux.HandleFunc(getApi(fmt.Sprintf(_getonething, _chirps, "{chirpID}")), func(w http.ResponseWriter, r *http.Request) {
		apipayload.GetUserRequestedChirp(w, r, apiCfg.db)
	})

	srv := &http.Server{
		Addr:    _defaultAddr,
		Handler: mux,
	}

	log.Printf("Seving on http://localhost%s\n", _defaultAddr)

	log.Fatal(srv.ListenAndServe())
}
