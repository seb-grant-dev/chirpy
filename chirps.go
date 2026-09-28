package main

import (
	"fmt"
	"time"
	"regexp"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/seb-grant-dev/chirpy/internal/database"
)

type ChirpError struct {
	error string
};

type Chirp struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type ChirpHandler struct {
	DB *database.Queries
}

func (h *ChirpHandler) getChirps(w http.ResponseWriter, req *http.Request) {

	chirps, err := h.DB.GetChirps(req.Context())


	if err != nil {
		respondWithError(w, 400, "Error retrieving chirps")
		return
	}

	var retChirps []Chirp

	for _, chirp := range chirps {
		retChirps = append(retChirps, Chirp(chirp))
	}

	respondWithJSON(w, 200, retChirps)
	return

}

func (h *ChirpHandler) getChirp(w http.ResponseWriter, req *http.Request) {

	chirpID,err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w,400,"Error parsing chirp ID from URL")
	}

	chirp, err := h.DB.GetChirp(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	respondWithJSON(w,200,Chirp(chirp))
	return
}


func (h *ChirpHandler) createChirp(w http.ResponseWriter, req *http.Request) {

	type params struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type response struct {
		chirp Chirp
	}

	var chirp *params
	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&chirp)

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error chirping: %s\n",err))
		return
	}

	err = validateChirp(chirp.Body)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error chirping: %s\n",err))
		return
	}



	chirpParams := database.CreateChirpParams{
		Body: cleanProfanity(chirp.Body),
		UserID: chirp.UserID,
	}


	newChirp, err := h.DB.CreateChirp(req.Context(), chirpParams)
	if err != nil {
		respondWithError(w,400,fmt.Sprintf("Error chirping: %s\n",err))
		return
	}

	fmt.Printf("%+v\n",newChirp)

	respondWithJSON(w,http.StatusCreated,Chirp{
			ID: newChirp.ID,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
			Body: newChirp.Body,
			UserID: newChirp.UserID,
		})

	return
}




func ReplaceAllCI(s, old, new string) string {
	// Compile the pattern with the (?i) flag for case-insensitivity
	// Use regexp.QuoteMeta to escape special regex characters in 'old'
	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(old))
	return pattern.ReplaceAllString(s, new)
}

func cleanProfanity(chirpBody string) string {
	blacklist := []string{"kerfuffle","sharbert","fornax"}

	for _, black := range blacklist {
		chirpBody = ReplaceAllCI(chirpBody,black,"****")
	}

	return chirpBody
}

func NewChirp(body string) *Chirp {
	return &Chirp{
		Body: body,
	}
}

func validateChirp(chirpBody string) error {

	if len(chirpBody) > 140 {
		return fmt.Errorf("Chirp is too long.")
	}

	return nil
}
