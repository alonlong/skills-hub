package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"skillhub/backend/internal/http/middleware"
	"skillhub/backend/internal/namespace"
)

type NamespaceHandler struct {
	service *namespace.Service
}

func NewNamespaceHandler(service *namespace.Service) *NamespaceHandler {
	return &NamespaceHandler{service: service}
}

func (h *NamespaceHandler) CreateNamespace(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input namespace.CreateNamespaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	created, err := h.service.CreateNamespace(r.Context(), actor, input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *NamespaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input namespace.AddMemberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	input.NamespaceSlug = chi.URLParam(r, "slug")

	if err := h.service.AddMember(r.Context(), actor, input); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *NamespaceHandler) GetNamespace(w http.ResponseWriter, r *http.Request) {
	namespaceData, err := h.service.GetNamespace(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, namespaceData)
}
