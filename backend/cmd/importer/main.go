package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/importer"
)

func main() {
	legacyDB := mustOpen(os.Getenv("LEGACY_DATABASE_URL"))
	defer legacyDB.Close()

	targetDB := mustOpen(os.Getenv("DATABASE_URL"))
	defer targetDB.Close()

	legacyPackageRoot := os.Getenv("LEGACY_PACKAGE_ROOT")
	if legacyPackageRoot == "" {
		log.Fatal("LEGACY_PACKAGE_ROOT is required")
	}

	packageRoot := os.Getenv("PACKAGE_ROOT")
	if packageRoot == "" {
		log.Fatal("PACKAGE_ROOT is required")
	}

	if err := importer.Run(legacyDB, targetDB, legacyPackageRoot, packageRoot); err != nil {
		log.Fatal(err)
	}
}

func mustOpen(dsn string) *sql.DB {
	if dsn == "" {
		log.Fatal("database URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}
