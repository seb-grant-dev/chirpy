package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/seb-grant-dev/chirpy/internal/database"
)

type User struct {
	ID 				uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email 		string 		`json:"email"`
}

type UserHandler struct {
	DB *database.Queries
}


func (h *UserHandler) createUser(w http.ResponseWriter, req *http.Request) {
	var user *User

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

	newUser,err := h.DB.CreateUser(req.Context(),user.Email)
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

