package uploads

import (
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ManagedRelativePath(raw string, publicPrefix string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	if parsed, err := url.Parse(trimmed); err == nil && parsed.Path != "" {
		trimmed = parsed.Path
	}

	cleaned := path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(trimmed), "/"))
	if !strings.HasPrefix(cleaned, publicPrefix) {
		return "", false
	}

	relative := strings.TrimPrefix(cleaned, publicPrefix)
	if relative == "" || relative == "." {
		return "", false
	}

	return relative, true
}

func PublicPath(publicPrefix string, storageKey string) string {
	trimmed := strings.TrimSpace(storageKey)
	if trimmed == "" {
		return ""
	}

	cleaned := path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(trimmed), "/"))
	return strings.TrimRight(publicPrefix, "/") + cleaned
}

func ResolvePath(uploadsDir string, storageKey string) (string, error) {
	root := filepath.Clean(strings.TrimSpace(uploadsDir))
	if root == "." || root == "" {
		return "", errors.New("uploads directory is required")
	}

	trimmed := strings.TrimSpace(storageKey)
	if trimmed == "" {
		return "", errors.New("storage key is required")
	}

	cleaned := path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(trimmed), "/"))
	relative := strings.TrimPrefix(cleaned, "/")
	if relative == "" || relative == "." {
		return "", errors.New("storage key is required")
	}

	candidate := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("storage key escapes uploads directory")
	}

	return candidate, nil
}

func Save(uploadsDir string, storageKey string, content []byte) error {
	target, err := ResolvePath(uploadsDir, storageKey)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, content, 0o644)
}

func Remove(uploadsDir string, storageKey string) error {
	target, err := ResolvePath(uploadsDir, storageKey)
	if err != nil {
		return err
	}

	return os.Remove(target)
}
