package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"skillhub/backend/internal/admin"
	"skillhub/backend/internal/http/middleware"
)

type ReviewHandler struct {
	service *admin.Service
}

func NewReviewHandler(service *admin.Service) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) Approve(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var request struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	if err := h.service.Approve(r.Context(), actor, chi.URLParam(r, "taskID"), request.Comment); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
