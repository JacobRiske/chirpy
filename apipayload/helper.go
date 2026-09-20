package apipayload

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func respondWithError(w http.ResponseWriter, code int, msg error) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg.Error(),
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

func vaildChirp(s string) (string, error) {

	const maxChirpLength = 140
	if len(s) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	return cleanBody(s), nil
}

func cleanBody(s string) string {
	var newBody string
	listWord := []string{}
	words := strings.Split(s, " ")
	for wl, w := range words {
		switch strings.ToLower(w) {
		case "kerfuffle", "sharbert", "fornax":
			if wl == len(words)-1 {
				listWord = append(listWord, "****")
			} else {
				listWord = append(listWord, "**** ")
			}
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
