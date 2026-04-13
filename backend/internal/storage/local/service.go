package local

import (
	"context"
	"os"
	"path/filepath"
)

type Service struct {
	basePath string
}

func NewService(basePath string) Service {
	return Service{basePath: basePath}
}

func (s Service) Save(ctx context.Context, key string, data []byte) (string, error) {
	_ = ctx
	fullPath := filepath.Join(s.basePath, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s Service) Read(ctx context.Context, key string) ([]byte, error) {
	_ = ctx
	return os.ReadFile(filepath.Join(s.basePath, key))
}
