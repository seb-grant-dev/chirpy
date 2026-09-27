package main

import (
	"net/http"
	"encoding/json"
)

type ChirpError struct {
	error string
};

type Chirp struct {
	Body string `json:"body"`
}

func NewChirp(body string) *Chirp {
	return &Chirp{
		Body: body,
	}
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	var chirp *Chirp

	type returnObj struct {
		Valid bool
	}

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&chirp)
	if err != nil {
		respondWithError(w, 400, "Error reading Chirp body")
		return
	}

	if chirp == nil {
		respondWithError(w, 400, "Error reading Chirp body")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp too long")
		return
	}

	respondWithJSON(w, http.StatusOK, returnObj{
		Valid:true,
	})
	return
}
