package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Cyberfly100/bootdev_httpserver/internal/auth"
	"github.com/Cyberfly100/bootdev_httpserver/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	userparams := struct {
		Email          string `json:"email"`
		HashedPassword string `json:"password"`
	}{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}
	r.Body.Close()

	err = json.Unmarshal(body, &userparams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	dbuserparams := database.CreateUserParams{
		Email:          userparams.Email,
		HashedPassword: userparams.HashedPassword,
	}

	dbuserparams.HashedPassword, err = auth.HashPassword(dbuserparams.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}

	var dbuser database.CreateUserRow
	dbuser, err = cfg.db.CreateUser(r.Context(), dbuserparams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	user := User{
		ID:        dbuser.ID,
		CreatedAt: dbuser.CreatedAt,
		UpdatedAt: dbuser.UpdatedAt,
		Email:     dbuser.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reset users", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users reset"))
}
