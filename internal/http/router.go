package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"skillhub/backend/internal/admin"
	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	"skillhub/backend/internal/http/handlers"
	appmiddleware "skillhub/backend/internal/http/middleware"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/search"
	"skillhub/backend/internal/skill"
)

type Dependencies struct {
	AuthService      *auth.Service
	NamespaceService *namespace.Service
	SkillService     *skill.Service
	AdminService     *admin.Service
	SearchService    *search.Service
}

// NewRouter builds the HTTP API router.
func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	_ = cfg

	r := chi.NewRouter()
	r.Get("/healthz", handlers.Health)

	if deps.AuthService != nil {
		authHandler := handlers.NewAuthHandler(deps.AuthService)
		r.Route("/api/v1/auth", func(authRouter chi.Router) {
			authRouter.Post("/login", authHandler.Login)
			authRouter.With(appmiddleware.RequireAuth(deps.AuthService)).Get("/me", authHandler.Me)
		})
	}

	if deps.AuthService != nil && deps.NamespaceService != nil {
		namespaceHandler := handlers.NewNamespaceHandler(deps.NamespaceService)
		r.With(appmiddleware.RequireAuth(deps.AuthService)).Post("/api/v1/namespaces", namespaceHandler.CreateNamespace)
		r.With(appmiddleware.RequireAuth(deps.AuthService)).Post("/api/v1/namespaces/{slug}/members", namespaceHandler.AddMember)
		r.Get("/api/v1/namespaces/{slug}", namespaceHandler.GetNamespace)
	}

	if deps.AuthService != nil && deps.SkillService != nil {
		skillHandler := handlers.NewSkillHandler(deps.SkillService, deps.SearchService)
		r.Get("/api/v1/skills/{namespace}/{skill}", skillHandler.GetSkill)
		r.Get("/api/v1/skills/{namespace}/{skill}/versions/{version}/download", skillHandler.DownloadVersion)
		r.With(appmiddleware.RequireAuth(deps.AuthService)).Post("/api/v1/skills/{namespace}/{skill}/versions", skillHandler.PublishVersion)
	}

	if deps.SearchService != nil {
		searchHandler := handlers.NewSkillHandler(deps.SkillService, deps.SearchService)
		r.Get("/api/v1/search", searchHandler.Search)
	}

	if deps.AuthService != nil && deps.AdminService != nil {
		reviewHandler := handlers.NewReviewHandler(deps.AdminService)
		r.With(appmiddleware.RequireAuth(deps.AuthService)).Post("/api/v1/reviews/{taskID}/approve", reviewHandler.Approve)
	}

	return r
}
