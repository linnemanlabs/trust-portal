// Package trustportal provides an HTTP handler that serves the files required for the trust portal, including the index page, CPS, trusted root configuration, signing configuration, and certificates.
package trustportal

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/linnemanlabs/go-core/log"
)

// Handler serves the trust portal files. It loads the files from disk at startup and serves them with appropriate content types.
type Handler struct {
	logger log.Logger
	files  map[string][]byte
}

// New creates a new Handler instance, loading the required files from the specified data directory. It panics if any file fails to load.
func New(logger log.Logger, dataDir string) *Handler {
	h := &Handler{
		logger: logger,
		files:  make(map[string][]byte),
	}

	entries := map[string]string{
		"index.html":                    "text/html",
		"favicon.svg":                   "image/svg+xml",
		"cps.html":                      "text/html",
		"trusted_root.json":             "application/json",
		"signing-config.json":           "application/json",
		"certs/root-ca.crt":             "application/x-pem-file",
		"certs/fulcio-ca.crt":           "application/x-pem-file",
		"certs/spire-ca.crt":            "application/x-pem-file",
		"certs/tsa.crt":                 "application/x-pem-file",
		"keys/rekor-checkpoint.pub":     "application/x-pem-file",
		"keys/tesseract-checkpoint.pub": "application/x-pem-file",
	}

	for filename := range entries {
		// Use filepath.Join and filepath.Clean to construct a safe file path (good practice, not needed on our own map of files but satisfies gosec linter)
		data, err := os.ReadFile(filepath.Clean(filepath.Join(dataDir, filename)))
		if err != nil {
			panic(fmt.Sprintf("failed to load %s: %v", filename, err))
		}
		h.files[filename] = data
	}

	return h
}

// RegisterRoutes registers the HTTP routes for the trust portal files on the given chi.Router. Each route serves a specific file with the appropriate content type.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.serveFile("index.html", "text/html"))
	r.Get("/favicon.svg", h.serveFile("favicon.svg", "image/svg+xml"))
	r.Get("/cps", h.serveFile("cps.html", "text/html"))

	r.Route("/.well-known", func(r chi.Router) {
		r.Get("/trusted_root.json", h.serveFile("trusted_root.json", "application/json"))
		r.Get("/signing-config.json", h.serveFile("signing-config.json", "application/json"))
	})

	r.Route("/certs", func(r chi.Router) {
		r.Get("/root-ca.crt", h.serveFile("certs/root-ca.crt", "application/x-pem-file"))
		r.Get("/fulcio-ca.crt", h.serveFile("certs/fulcio-ca.crt", "application/x-pem-file"))
		r.Get("/spire-ca.crt", h.serveFile("certs/spire-ca.crt", "application/x-pem-file"))
		r.Get("/tsa.crt", h.serveFile("certs/tsa.crt", "application/x-pem-file"))
	})

	r.Route("/keys", func(r chi.Router) {
		r.Get("/rekor-checkpoint.pub", h.serveFile("keys/rekor-checkpoint.pub", "application/x-pem-file"))
		r.Get("/tesseract-checkpoint.pub", h.serveFile("keys/tesseract-checkpoint.pub", "application/x-pem-file"))
	})
}

// serveFile returns an http.HandlerFunc that serves the specified file with the given content type. If the file is not found in the handler's files map, it responds with a 404 error.
func (h *Handler) serveFile(filename, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		data, ok := h.files[filename]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(data)
	}
}
