package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"skillhub/backend/internal/auth"
)

type Storage interface {
	Save(ctx context.Context, key string, data []byte) (string, error)
	Read(ctx context.Context, key string) ([]byte, error)
}

type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func ErrSkillNotFound() error {
	return errSkillNotFound
}

func ErrNamespaceRoleMissing() error {
	return errNamespaceRoleMissing
}

func (s *Service) PublishVersion(ctx context.Context, actor auth.Actor, input PublishInput) (Version, error) {
	role, err := s.repo.NamespaceRole(ctx, input.NamespaceSlug, actor.UserID)
	if err != nil {
		return Version{}, err
	}
	if role != "OWNER" && role != "ADMIN" && role != "MEMBER" {
		return Version{}, errors.New("forbidden")
	}

	meta, err := parseManifest(input.Archive)
	if err != nil {
		return Version{}, err
	}

	storageKey := filepath.Join(input.NamespaceSlug, input.SkillSlug, meta.Version+".zip")
	storagePath, err := s.storage.Save(ctx, storageKey, input.Archive)
	if err != nil {
		return Version{}, err
	}

	manifestJSON, err := json.Marshal(meta)
	if err != nil {
		return Version{}, err
	}

	return s.repo.CreateVersion(ctx, CreateVersionParams{
		NamespaceSlug: input.NamespaceSlug,
		SkillSlug:     input.SkillSlug,
		Version:       meta.Version,
		DisplayName:   meta.DisplayName,
		Summary:       meta.Summary,
		Status:        "PENDING_REVIEW",
		ManifestJSON:  string(manifestJSON),
		MetadataJSON:  string(manifestJSON),
		StoragePath:   storagePath,
		SubmittedBy:   actor.UserID,
	})
}

func (s *Service) DownloadVersion(ctx context.Context, namespaceSlug string, skillSlug string, version string) ([]byte, error) {
	storagePath, err := s.repo.VersionStorage(ctx, namespaceSlug, skillSlug, version)
	if err != nil {
		return nil, err
	}
	key := filepath.Join(namespaceSlug, skillSlug, version+".zip")
	if strings.HasSuffix(storagePath, key) {
		return s.storage.Read(ctx, key)
	}
	return s.storage.Read(ctx, key)
}

func parseManifest(archive []byte) (manifest, error) {
	readerAt := bytes.NewReader(archive)
	zipReader, err := zip.NewReader(readerAt, int64(len(archive)))
	if err != nil {
		return manifest{}, err
	}

	for _, file := range zipReader.File {
		if file.Name != "manifest.json" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return manifest{}, err
		}
		defer rc.Close()

		payload, err := io.ReadAll(rc)
		if err != nil {
			return manifest{}, err
		}

		var meta manifest
		if err := json.Unmarshal(payload, &meta); err != nil {
			return manifest{}, err
		}
		if meta.Version == "" {
			return manifest{}, fmt.Errorf("manifest version is required")
		}
		if meta.DisplayName == "" {
			meta.DisplayName = "Unnamed Skill"
		}
		return meta, nil
	}

	return manifest{}, fmt.Errorf("manifest.json not found")
}
