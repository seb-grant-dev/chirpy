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

func (h *UserHandler) doLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
		ExpiresInSeconds int `json:"expires_in_seconds"`
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

		if user.ExpiresInSeconds != 0 {
			if (user.ExpiresInSeconds < expirySecs) {
				expirySecs = user.ExpiresInSeconds
			}
		}

		
		fmt.Println(expirySecs)

		expiry := time.Duration(expirySecs) * time.Second

		token, err := auth.MakeJWT(loggedInUser.ID, h.jwtSecret, expiry)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Login failed: %s\n",err))
			return
		}

		fmt.Sprintf("Token: %s",token)

		retUser := User{
			ID: loggedInUser.ID,
			CreatedAt: loggedInUser.CreatedAt,
			UpdatedAt: loggedInUser.UpdatedAt,
			Email: loggedInUser.Email,
			Token: token,
		}

		respondWithJSON(w, http.StatusOK, retUser)
		return
	} else {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
}
