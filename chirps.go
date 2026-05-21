package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Cyberfly100/bootdev_httpserver/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func validateChirp(chirpParams *database.CreateChirpParams) error {
	const maxChirpLength = 140

	if len(chirpParams.Body) == 0 {
		return fmt.Errorf("Chirp body cannot be empty")
	}

	if len(chirpParams.Body) > maxChirpLength {
		return fmt.Errorf("Chirp body cannot exceed %d characters", maxChirpLength)
	}

	const censorstring = "****"
	filterProfanity(&chirpParams.Body, censorstring)
	return nil
}

func filterProfanity(body *string, censor string) {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, profanity := range profaneWords {
		for word := range strings.SplitSeq(*body, " ") {
			if strings.EqualFold(word, profanity) {
				*body = strings.ReplaceAll(*body, word, censor)
			}
		}
	}
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type ChirpParams struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	var chirpParams ChirpParams
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}
	r.Body.Close()

	err = json.Unmarshal(body, &chirpParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	dbchirpParams := database.CreateChirpParams{
		Body:   chirpParams.Body,
		UserID: chirpParams.UserID,
	}

	err = validateChirp(&dbchirpParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var dbchirp database.Chirp
	dbchirp, err = cfg.db.CreateChirp(r.Context(), dbchirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	chirp := Chirp{
		ID:        dbchirp.ID,
		CreatedAt: dbchirp.CreatedAt,
		UpdatedAt: dbchirp.UpdatedAt,
		Body:      dbchirp.Body,
		UserID:    dbchirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}
