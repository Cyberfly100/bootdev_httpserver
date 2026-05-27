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

type LoginParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	accessTokenExpiration := time.Hour
	authToken, err := auth.MakeJWT(user.ID, cgf.JWTSecret, accessTokenExpiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT", err)
		return
	}

	refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 60)

	refreshToken := auth.MakeRefreshToken()

	refreshTokenParams := database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshTokenExpiresAt,
	}

	refreshTokenInfo, err := cgf.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	returnUser := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token,omitempty"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        authToken,
		RefreshToken: refreshTokenInfo.Token,
	}

	respondWithJSON(w, http.StatusOK, returnUser)
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header", err)
		return
	}

	var refreshTokenInfo database.RefreshToken
	refreshTokenInfo, err = cfg.db.GetRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	if refreshTokenInfo.RevokedAt.Valid || refreshTokenInfo.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is expired or revoked", nil)
		return
	}

	accessToken, err := auth.MakeJWT(refreshTokenInfo.UserID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) validateAccessToken(r *http.Request) (uuid.UUID, error) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}
