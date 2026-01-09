package handlers

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/storage"
)

type AssetHandler struct {
	provider storage.Provider
}

func NewAssetHandler(provider storage.Provider) *AssetHandler {
	return &AssetHandler{
		provider: provider,
	}
}

func (h *AssetHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/assets/")

	if path == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	cleanPath := filepath.Clean(path)

	if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") || strings.Contains(path, ":") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	reader, err := h.provider.GetObject(r.Context(), cleanPath)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	contentType := detectContentType(cleanPath)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if r.Method == http.MethodHead {
		return
	}

	if _, err := io.Copy(w, reader); err != nil {
		return
	}
}

func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	types := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}

	if ct, ok := types[ext]; ok {
		return ct
	}

	return "application/octet-stream"
}
