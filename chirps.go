package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Cyberfly100/bootdev_httpserver/internal/auth"
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
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	type ChirpParams struct {
		Body string `json:"body"`
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
		UserID: userID,
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

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	const defaultLimit = 100

	var dbchirps []database.Chirp

	sorting_option := r.URL.Query().Get("sort")
	switch sorting_option {
	case "asc", "desc":
	// valid sorting options, do nothing
	default:
		sorting_option = "asc"
	}

	authorIDStr := r.URL.Query().Get("author_id")
	if authorIDStr != "" {
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id format", err)
			return
		}

		getChirpsByUserIDParams := database.GetChirpsByUserIDParams{
			UserID: authorID,
			Limit:  defaultLimit,
		}
		dbchirps, err = cfg.db.GetChirpsByUserID(r.Context(), getChirpsByUserIDParams)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps", err)
			return
		}
	} else {
		var err error
		dbchirps, err = cfg.db.GetChirps(r.Context(), defaultLimit)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps", err)
			return
		}
	}
	chirps := make([]Chirp, len(dbchirps))
	for i, dbchirp := range dbchirps {
		chirps[i] = Chirp{
			ID:        dbchirp.ID,
			CreatedAt: dbchirp.CreatedAt,
			UpdatedAt: dbchirp.UpdatedAt,
			Body:      dbchirp.Body,
			UserID:    dbchirp.UserID,
		}
	}
	if sorting_option == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID", nil)
		return
	}

	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	dbchirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to retrieve chirp", err)
		return
	}

	chirp := Chirp{
		ID:        dbchirp.ID,
		CreatedAt: dbchirp.CreatedAt,
		UpdatedAt: dbchirp.UpdatedAt,
		Body:      dbchirp.Body,
		UserID:    dbchirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.validateAccessToken(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid access token", err)
		return
	}

	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID", nil)
		return
	}
	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	dbchirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if dbchirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
