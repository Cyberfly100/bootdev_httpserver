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

type LoginParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func (cgf *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}
	r.Body.Close()

	var loginParams LoginParams

	err = json.Unmarshal(body, &loginParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	user, err := cgf.db.GetUserByEmail(r.Context(), loginParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	match, err := auth.CheckPasswordHash(loginParams.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	expirationTime := time.Hour
	if loginParams.ExpiresInSeconds > 0 && loginParams.ExpiresInSeconds < int(expirationTime/time.Second) {
		expirationTime = time.Duration(loginParams.ExpiresInSeconds) * time.Second
	}

	authToken, err := auth.MakeJWT(user.ID, cgf.JWTSecret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT", err)
		return
	}

	returnUser := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     authToken,
	}

	respondWithJSON(w, http.StatusOK, returnUser)
}
