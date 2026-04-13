package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"skillhub/backend/internal/http/middleware"
	"skillhub/backend/internal/search"
	"skillhub/backend/internal/skill"
)

type SkillHandler struct {
	service       *skill.Service
	searchService *search.Service
}

func NewSkillHandler(service *skill.Service, searchService *search.Service) *SkillHandler {
	return &SkillHandler{
		service:       service,
		searchService: searchService,
	}
}

func (h *SkillHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	actor, ok := middleware.ActorFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_file"})
		return
	}
	defer file.Close()

	archive, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_file"})
		return
	}

	version, err := h.service.PublishVersion(r.Context(), actor, skill.PublishInput{
		NamespaceSlug: chi.URLParam(r, "namespace"),
		SkillSlug:     chi.URLParam(r, "skill"),
		Archive:       archive,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, version)
}

func (h *SkillHandler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	payload, err := h.service.DownloadVersion(
		r.Context(),
		chi.URLParam(r, "namespace"),
		chi.URLParam(r, "skill"),
		chi.URLParam(r, "version"),
	)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (h *SkillHandler) GetSkill(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not_implemented"})
}

func (h *SkillHandler) Search(w http.ResponseWriter, r *http.Request) {
	results, err := h.searchService.Search(r.Context(), search.Query{
		Text:          r.URL.Query().Get("q"),
		NamespaceSlug: r.URL.Query().Get("namespace"),
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_search"})
		return
	}
	writeJSON(w, http.StatusOK, results)
}
