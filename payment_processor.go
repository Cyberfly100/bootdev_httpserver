package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}
	r.Body.Close()
	type PolkaWebhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var webhookReq PolkaWebhookRequest
	err = json.Unmarshal(body, &webhookReq)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	if webhookReq.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Unsupported event type", nil)
		return
	}

	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), webhookReq.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to upgrade user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
