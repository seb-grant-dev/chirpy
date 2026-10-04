package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/seb-grant-dev/chirpy/internal/database"
)

type WebhookParams struct {
	Event string `json:"event"`
	Data struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

type WebhookHandler struct {
	DB *database.Queries
	jwtSecret string
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

	fmt.Printf("%\n",parameters.Event)

	switch parameters.Event {
	case "user.upgraded":
		h.upgradeUser(w, req, parameters)
	default:
		fmt.Printf("Params: %+v\n",parameters)
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}
}

func (h *WebhookHandler) upgradeUser(w http.ResponseWriter, req *http.Request, params *WebhookParams) {

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
