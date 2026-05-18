package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type chirp struct {
	Body string `json:"body"`
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var c chirp
	err := decoder.Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode chirps JSON", err)
		return
	}

	validateChirp(w, c)
}

func validateChirp(w http.ResponseWriter, c chirp) {
	const maxChirpLength = 140

	type chirpReply struct {
		Valid bool `json:"valid"`
	}

	if len(c.Body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Chirp body cannot be empty", nil)
		return
	}

	if len(c.Body) > maxChirpLength {
		msg := fmt.Sprintf("Chirp body cannot exceed %d characters", maxChirpLength)
		respondWithError(w, http.StatusBadRequest, msg, nil)
		return
	}

	respondWithJSON(w, http.StatusOK, chirpReply{Valid: true})
}
