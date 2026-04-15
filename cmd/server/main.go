package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/admin"
	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	"skillhub/backend/internal/label"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/search"
	"skillhub/backend/internal/skill"
	localstorage "skillhub/backend/internal/storage/local"
	"skillhub/backend/internal/store/postgres"
)

func main() {
	cfg := config.Load()
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	handler, err := newHandler(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + cfg.HTTPPort
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Printf("listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func newHandler(cfg config.Config, db *sql.DB) (http.Handler, error) {
	authRepo := auth.NewSQLRepository(db)
	authService := auth.NewService(authRepo, auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	if err := authService.EnsureBootstrapAdmin(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure bootstrap admin: %w", err)
	}

	namespaceRepo := namespace.NewSQLRepository(db)
	namespaceService := namespace.NewService(namespaceRepo)

	searchRepo := postgres.NewSearchRepository(db)
	searchService := search.NewService(searchRepo)

	storageService := localstorage.NewService(cfg.StorageBasePath)
	skillRepo := postgres.NewSkillRepository(db)
	skillService := skill.NewService(skillRepo, storageService)

	reviewRepo := postgres.NewReviewRepository(db)
	adminService := admin.NewService(reviewRepo)

	labelRepo := postgres.NewLabelRepository(db)
	labelService := label.NewService(labelRepo)

	return aphttp.NewRouter(cfg, aphttp.Dependencies{
		AuthService:      authService,
		NamespaceService: namespaceService,
		SkillService:     skillService,
		AdminService:     adminService,
		SearchService:    searchService,
		LabelService:     labelService,
	}), nil
}
