package apipayload

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/JacobRiske/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func CreateChirp(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		chirped    database.Chirp
		cleanChirp string
		err        error
	)

	if db == nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("Database is not configured"))
		return
	}

	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid user_id"))
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

func AddUser(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		err        error
		addNewUser database.User
	)

	if db == nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("Database is not configured"))
		return
	}

	type parameters struct {
		Email string `json:"email"`
	}

	type returnVals struct {
		ID     uuid.UUID `json:"id"`
		Create time.Time `json:"created_at"`
		Update time.Time `json:"updated_at"`
		Email  string    `json:"email"`
	}

	data := json.NewDecoder(r.Body)
	params := parameters{}
	if err = data.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	if addNewUser, err = db.CreateUser(context.Background(), params.Email); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to add the new user"))
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		ID:     addNewUser.ID,
		Create: addNewUser.CreatedAt,
		Update: addNewUser.UpdatedAt,
		Email:  addNewUser.Email,
	})

}

func ReturnAllUserChirps(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		userChirps          []database.Chirp
		returnAllUserChirps []Chirp
		id                  uuid.UUID
		err                 error
	)

	userID := r.PathValue("userID")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	id, err = uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid user id"))
		return
	}

	if userChirps, err = db.GetAllUserChirp(context.Background(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get users chirps"))
		return
	}

	for _, chirp := range userChirps {
		returnAllUserChirps = append(returnAllUserChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, userChirps)
}

func GetAllChirps(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		allChirps       []database.Chirp
		returnAllChirps []Chirp
		err             error
	)

	if allChirps, err = db.GetAllChirps(context.Background()); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get chirps"))
		return
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
