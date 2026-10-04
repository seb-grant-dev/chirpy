package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/seb-grant-dev/chirpy/internal/database"
	"github.com/seb-grant-dev/chirpy/internal/auth"
)

type User struct {
	ID 				uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email 		string 		`json:"email"`
	Token			string 		`json:"token"`
	RefreshToken			string 		`json:"refresh_token"`
}

type UserHandler struct {
	DB *database.Queries
	jwtSecret string
}


func (h *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {

	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	var user *parameters

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, 400, "Error reading User body")
		return
	}

	if user == nil {
		respondWithError(w, 400, "Error reading User body")
		return
	}

	hashed,err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, 400, "Error creating password")
	}

	newUserParams := database.CreateUserParams{
		Email: user.Email,
		HashedPassword: hashed,
	}

	newUser,err := h.DB.CreateUser(req.Context(),newUserParams)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error creating user: %s",err))
		return
	}

	retUser := User{
		ID: newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, retUser)


	return
}

func (h *UserHandler) updateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	var user *parameters
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters")
		return
	}

	accessToken,err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid access token.")
		return
	}

	userId, err := auth.ValidateJWT(accessToken, h.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect user ID")
		return
	}

	hashedPassword, _ := auth.HashPassword(user.Password)

	updateUserParams := database.UpdateUserParams{
		ID: userId,
		HashedPassword: hashedPassword,
		Email: user.Email,
	}


	updatedUser, err := h.DB.UpdateUser(req.Context(),updateUserParams)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Could not update user. %s",err))
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID: updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email: updatedUser.Email,
	})
	return
}

func (h *UserHandler) doLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	var user *parameters
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s\n",err))
		return
	}

	loggedInUser, err := h.DB.GetUserForEmail(req.Context(),user.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("No user found for those credentials: %s",err))
		return
	}

	result, err := auth.CheckPasswordHash(user.Password, loggedInUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	}

	if result {

		expirySecs := 3600
		expiry := time.Duration(expirySecs) * time.Second

		token, err := auth.MakeJWT(loggedInUser.ID, h.jwtSecret, expiry)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s\n",err))
			return
		}

		refresh_token := auth.MakeRefreshToken()
		if refresh_token == "" {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Login failed: Refresh token not generated\n"))
		}

		refreshTokenParams := database.SaveUserRefreshTokenParams{
			Token: refresh_token,
			UserID: loggedInUser.ID,
			ExpiresAt: time.Now().AddDate(0,0,60),
		}
		_,err = h.DB.SaveUserRefreshToken(req.Context(),refreshTokenParams)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error saving refresh_token: %s\n",err))
		}

		retUser := User{
			ID: loggedInUser.ID,
			CreatedAt: loggedInUser.CreatedAt,
			UpdatedAt: loggedInUser.UpdatedAt,
			Email: loggedInUser.Email,
			Token: token,
			RefreshToken: refresh_token,
		}

		respondWithJSON(w, http.StatusOK, retUser)
		return
	} else {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
}

func (h *UserHandler) refreshToken(w http.ResponseWriter, req *http.Request) {
	token,err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error refreshing tokens")
		return
	}

	refresh_token, err := h.DB.GetRefreshToken(req.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	if refresh_token.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	type retParam struct {
		Token string `json:"token"`
	}


	// Get a new JWT
	expirySecs := 3600
	expiry := time.Duration(expirySecs) * time.Second
	new_token, err := auth.MakeJWT(refresh_token.UserID, h.jwtSecret, expiry)


	respondWithJSON(w, http.StatusOK, retParam{
		Token: new_token,
	})
	return

}

func (h *UserHandler) revokeToken(w http.ResponseWriter, req *http.Request) {
	token,err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error refreshing tokens")
		return
	}

	
	err = h.DB.RevokeToken(req.Context(),token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
	return

}
