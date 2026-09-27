package main

import (
	"regexp"
	"net/http"
	"encoding/json"
)

type ChirpError struct {
	error string
};

type Chirp struct {
	Body string `json:"body"`
}
func ReplaceAllCI(s, old, new string) string {
	// Compile the pattern with the (?i) flag for case-insensitivity
	// Use regexp.QuoteMeta to escape special regex characters in 'old'
	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(old))
	return pattern.ReplaceAllString(s, new)
}

func (c *Chirp) cleanProfanity() {
	blacklist := []string{"kerfuffle","sharbert","fornax"}
	for _, black := range blacklist {
		c.Body = ReplaceAllCI(c.Body,black,"****")
	}
}

func NewChirp(body string) *Chirp {
	return &Chirp{
		Body: body,
	}
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	var chirp *Chirp

	type returnChirp struct {
		Body string `json:"cleaned_body"`
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

	chirp.cleanProfanity()

	respondWithJSON(w, http.StatusOK, returnChirp{
		Body: chirp.Body,
	})

	return
}
