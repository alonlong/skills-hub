package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"skillhub/backend/internal/auth"
	appmiddleware "skillhub/backend/internal/http/middleware"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	token, user, err := h.service.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken": token,
		"tokenType":   "Bearer",
		"user":        user,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	actor, ok := appmiddleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.service.CurrentUser(r.Context(), actor)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	actor, ok := appmiddleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var request struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	err := h.service.ChangePassword(r.Context(), actor, request.CurrentPassword, request.NewPassword)
	if errors.Is(err, auth.ErrChangePasswordInvalidCurrent) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}
	if errors.Is(err, auth.ErrChangePasswordWeak) || errors.Is(err, auth.ErrChangePasswordInvalidInput) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
