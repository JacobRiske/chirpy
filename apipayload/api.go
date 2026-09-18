package apipayload

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func ChirpValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Body: cleanBody(params.Body),
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func cleanBody(s string) string {
	var newBody string
	listWord := []string{}
	words := strings.Split(s, " ")
	for wl, w := range words {
		switch w {
		case "kerfuffle", "sharbert", "fornax", "Fornax", "Kerfuffle", "Sharbert":
			listWord = append(listWord, "**** ")
		default:
			if wl == len(words)-1 {
				listWord = append(listWord, w)
			} else {
				listWord = append(listWord, w+" ")
			}
		}
	}

	return strings.Join(listWord, newBody)
}
