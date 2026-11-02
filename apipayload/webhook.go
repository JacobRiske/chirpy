package apipayload

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/JacobRiske/chirpy/internal/auth"
	"github.com/JacobRiske/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func UpdateChiryRed(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	type DataPoint struct {
		UserID string `json:"user_id"`
	}

	type parameters struct {
		Event string    `json:"event"`
		Data  DataPoint `json:"data"`
	}

	data := json.NewDecoder(r.Body)
	params := parameters{}
	if err := data.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	if _, err := auth.GetApiToken(r.Header); err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "Un-authorized access"))
		return
	}

	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.Wrap(err, "invalid user id"))
		return
	}

	db.UpdateUserChirpyRed(context.Background(), database.UpdateUserChirpyRedParams{
		ID:          id,
		IsChirpyRed: true,
	})

	respondWithJSON(w, http.StatusNoContent, nil)

}
