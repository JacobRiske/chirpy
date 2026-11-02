package apipayload

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/JacobRiske/chirpy/internal/auth"
	"github.com/JacobRiske/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func AddUser(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	var (
		err        error
		addNewUser database.User
		hashedPass string
	)

	if db == nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("Database is not configured"))
		return
	}

	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
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

	if hashedPass, err = auth.HashPass(params.Pass); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to hash password"))
		return
	}

	if addNewUser, err = db.CreateUser(context.Background(), database.CreateUserParams{Email: params.Email, Pass: hashedPass}); err != nil {
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

func Login(w http.ResponseWriter, r *http.Request, db *database.Queries, sToken string) {
	var (
		err      error
		userData database.User
		ok       bool
		token    string
	)

	type parameters struct {
		Email    string  `json:"email"`
		Pass     string  `json:"password"`
		TokenEpx float64 `json:"expires_in_second"`
	}

	type returnVals struct {
		ID            uuid.UUID `json:"id"`
		Create        time.Time `json:"created_at"`
		Update        time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Token         string    `json:"token"`
		RefreashToken string    `json:"refresh_token"`
	}

	data := json.NewDecoder(r.Body)
	params := parameters{}
	if err = data.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	if userData, err = db.LookUpUserPass(context.Background(), params.Email); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to get users data"))
		return
	}

	tokenData, err := db.LookUpUserRefeashToken(r.Context(), userData.ID)
	rTok := tokenData.Token
	if errors.Is(err, sql.ErrNoRows) {
		created, createErr := db.CreateUserRefeashToken(
			r.Context(),
			database.CreateUserRefeashTokenParams{
				Token:  auth.MakeRefreshToken(),
				UserID: userData.ID,
				ExpiresAt: sql.NullTime{
					Time:  time.Now().AddDate(0, 0, 60),
					Valid: true,
				},
				RevokedAt: sql.NullTime{Valid: false},
			},
		)
		if createErr != nil {
			respondWithError(w, http.StatusInternalServerError, errors.Wrap(createErr, "Failed to create refresh token"))
			return
		}
		rTok = created.Token
	} else if tokenData.RevokedAt.Valid {
		updateToken, updateErr := db.ResetRefeashToken(context.Background(),
			database.ResetRefeashTokenParams{
				Token:     auth.MakeRefreshToken(),
				RevokedAt: sql.NullTime{Valid: false},
				UserID:    userData.ID,
			})
		if updateErr != nil {
			respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to update refresh token"))
			return
		}
		rTok = updateToken.Token
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to look up refresh token"))
		return
	}

	if ok, err = auth.CheckPass(params.Pass, userData.Pass); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to check auth on inputs"))
		return
	}

	if !ok {
		respondWithError(w, http.StatusUnauthorized, errors.New("User is unauthorized"))
		return
	}

	expiration := 3600 * time.Second

	if params.TokenEpx > 0 {
		expiration = time.Duration(params.TokenEpx) * time.Second
	}

	if token, err = auth.MakeJWT(userData.ID, sToken, expiration); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed make token"))
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		ID:            userData.ID,
		Create:        userData.CreatedAt,
		Update:        userData.UpdatedAt,
		Email:         userData.Email,
		Token:         token,
		RefreashToken: rTok,
	})

}

func RefreshToken(w http.ResponseWriter, r *http.Request, db *database.Queries, sToken string) {
	var (
		token, rtoken string
		err           error
		tokenStatus   database.CheckTokenRow
	)

	type refreshedToken struct {
		Token string `json:"token"`
	}

	if rtoken, err = auth.GetBearToken(r.Header); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed check token"))
		return
	}

	if tokenStatus, err = db.CheckToken(context.Background(), rtoken); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to pull refresh token data"))
		return
	}

	if tokenStatus.RevokedAt.Valid || !tokenStatus.ExpiresAt.Time.After(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, errors.New("User token is un-authorized"))
		return
	}

	if token, err = auth.MakeJWT(tokenStatus.UserID, sToken, expiration); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to make new token"))
		return
	}

	respondWithJSON(w, http.StatusOK, refreshedToken{
		Token: token,
	})
}

func RevokeToken(w http.ResponseWriter, r *http.Request, db *database.Queries) {
	rtoken, err := auth.GetBearToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed check token"))
		return
	}

	_, err = db.RevokedToken(context.Background(), rtoken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed revoke token"))
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func UpdateUserData(w http.ResponseWriter, r *http.Request, db *database.Queries, stoken string) {

	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
	}

	type returnData struct {
		Email string `json:"email"`
	}

	userToken, err := auth.GetBearToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "un-authorized user"))
		return
	}

	data := json.NewDecoder(r.Body)
	params := parameters{}
	if err = data.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Couldn't decode parameters"))
		return
	}

	userID, err := auth.ValidateJWT(userToken, stoken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, errors.Wrap(err, "un-authorized user"))
		return
	}

	hashedPass, err := auth.HashPass(params.Pass)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to hash password"))
		return
	}

	updatedData, err := db.UpdateUserData(context.Background(),
		database.UpdateUserDataParams{
			Email: params.Email,
			Pass:  hashedPass,
			ID:    userID,
		})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "Failed to update user data"))
		return
	}

	respondWithJSON(w, http.StatusOK, returnData{
		Email: updatedData.Email,
	})
}
