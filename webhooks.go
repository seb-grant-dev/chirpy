package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/seb-grant-dev/chirpy/internal/database"
	"github.com/seb-grant-dev/chirpy/internal/auth"
)

type WebhookParams struct {
	Event string `json:"event"`
	Data struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

type WebhookHandler struct {
	DB *database.Queries
	polkaKey string
}

func (h *WebhookHandler) handle(w http.ResponseWriter, req *http.Request) {

	// Decode params
	var parameters *WebhookParams
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&parameters)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed webhook body")
		return
	}

	switch parameters.Event {
	case "user.upgraded":
		h.upgradeUser(w, req, parameters)
	default:
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}
}

func (h *WebhookHandler) upgradeUser(w http.ResponseWriter, req *http.Request, params *WebhookParams) {

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized,"Error: API Key not provided")
		return
	}

	if apiKey != h.polkaKey {
		respondWithError(w,http.StatusUnauthorized,"Error: Invalid API Key")
		return
	}


	user, err := h.DB.GetUser(req.Context(),params.Data.UserID)
	fmt.Printf("\nParams: %+v\nUser: %+v\n",params,user)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	_,err = h.DB.UpgradeUser(req.Context(),user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error upgrading account: %s",err))
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
	return
}
