package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
		CleanedBody string `json:"cleaned_body"`
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

	const censorstring = "****"
	filterProfanity(&c.Body, censorstring)

	respondWithJSON(w, http.StatusOK, chirpReply{CleanedBody: c.Body})
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
