package main

import (
	"net/http"
	"encoding/json"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
	return
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Add("Content-Type","application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		// todo
	}
	w.WriteHeader(code)
	w.Write(data)
	return
}

