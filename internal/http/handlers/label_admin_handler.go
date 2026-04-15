package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"skillhub/backend/internal/http/middleware"
	"skillhub/backend/internal/label"
)

type LabelAdminHandler struct {
	svc *label.Service
}

func NewLabelAdminHandler(svc *label.Service) *LabelAdminHandler {
	return &LabelAdminHandler{svc: svc}
}

func (h *LabelAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	defs, err := h.svc.List(r.Context(), actor)
	if err != nil {
		writeLabelError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, defs)
}

func (h *LabelAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var in label.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	def, err := h.svc.Create(r.Context(), actor, in)
	if err != nil {
		writeLabelError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, def)
}

func (h *LabelAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	slug := chi.URLParam(r, "slug")
	var in label.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	def, err := h.svc.Update(r.Context(), actor, slug, in)
	if err != nil {
		writeLabelError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, def)
}

func (h *LabelAdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	slug := chi.URLParam(r, "slug")
	if err := h.svc.Delete(r.Context(), actor, slug); err != nil {
		writeLabelError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelAdminHandler) SortOrder(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req struct {
		Items []label.SortOrderItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	defs, err := h.svc.UpdateSortOrder(r.Context(), actor, req.Items)
	if err != nil {
		writeLabelError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, defs)
}

func writeLabelError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, label.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, label.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
	case errors.Is(err, label.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "slug_exists"})
	case errors.Is(err, label.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid"})
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}
