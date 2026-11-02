package apipayload

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/JacobRiske/chirpy/internal/auth"
	"github.com/JacobRiske/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func CreateChirp(w http.ResponseWriter, r *http.Request, db *database.Queries, sToken string) {
	var (
		chirped           database.Chirp
		cleanChirp, token string
		err               error
	)

	if db == nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("Database is not configured"))
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	if token, err = auth.GetBearToken(r.Header); err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "invalid user infromation"))
		return
	}

	userID, err := auth.ValidateJWT(token, sToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "invalid user_id"))
		return
	}

	if cleanChirp, err = vaildChirp(params.Body); err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "failed to vaildate the chrip"))
		return
	}

	readyChirp := database.CreateChirpParams{
		Body:   cleanChirp,
		UserID: userID,
	}

	if chirped, err = db.CreateChirp(context.Background(), readyChirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to post chirp"))
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirped.ID,
		Body:      chirped.Body,
		CreatedAt: chirped.CreatedAt,
		UpdatedAt: chirped.UpdatedAt,
		UserID:    chirped.UserID,
	})
}

func GetAllChirps(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		allChirps       []database.Chirp
		returnAllChirps []Chirp
		err             error
	)

	s := r.URL.Query().Get("author_id")

	if s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid user id"))
			return
		}
		switch r.URL.Query().Get("sort") {
		case "asc":
			if allChirps, err = db.GetAllUserChirpASC(context.Background(), id); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		case "desc":
			if allChirps, err = db.GetAllUserChirpDESC(context.Background(), id); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		default:
			if allChirps, err = db.GetAllUserChirp(context.Background(), id); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		}
	} else {
		switch r.URL.Query().Get("sort") {
		case "asc":
			if allChirps, err = db.GetAllChirpsASC(context.Background()); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		case "desc":
			if allChirps, err = db.GetAllChirpsDESC(context.Background()); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		default:
			if allChirps, err = db.GetAllChirps(context.Background()); err != nil {
				respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
				return
			}
		}
	}

	for _, chirp := range allChirps {
		returnAllChirps = append(returnAllChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, returnAllChirps)
}

func GetUserRequestedChirp(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("missing chirp id"))
		return
	}

	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid chirp id"))
		return
	}

	chirpData, err := db.GetOneChirp(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, errors.New("chirp not found"))
			return
		}
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get requested chirp"))
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirpData.ID,
		CreatedAt: chirpData.CreatedAt,
		UpdatedAt: chirpData.UpdatedAt,
		Body:      chirpData.Body,
		UserID:    chirpData.UserID,
	})

}

func RemoveChirp(w http.ResponseWriter, r *http.Request, db *database.Queries, stoken string) {

	chirpID := r.PathValue("chirpID")

	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("missing chirp id"))
		return
	}

	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid chirp id"))
		return
	}

	token, err := auth.GetBearToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "Un-authorized token"))
		return
	}

	userID, err := auth.ValidateJWT(token, stoken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "Un-authorized user"))
		return
	}

	chirpData, err := db.GetOneChirp(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, errors.Wrap(err, "missing chirp"))
			return
		}
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get chirp data"))
		return
	}

	if chirpData.UserID != userID {
		respondWithError(w, http.StatusForbidden, errors.New("Un-authorized user request"))
		return
	}

	_, err = db.RemoveChirp(context.Background(), database.RemoveChirpParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, errors.Wrap(err, "missing chirp"))
			return
		}
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to remove chirp"))
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
