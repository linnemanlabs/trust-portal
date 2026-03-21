package trustportal

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
	"github.com/linnemanlabs/go-core/log"
)

// testLogger returns a logger that discards output.
func testLogger(t *testing.T) log.Logger {
	t.Helper()
	lg, err := log.New(&log.Options{App: "test"})
	if err != nil {
		t.Fatalf("log.New: %v", err)
	}
	return lg
}

// setupTestData returns an in-memory fs.FS with all required trust portal files.
func setupTestData(t *testing.T) fs.FS {
	t.Helper()
	return fstest.MapFS{
		"index.html":                    {Data: []byte("<html><body>index</body></html>")},
		"favicon.svg":                   {Data: []byte("<svg></svg>")},
		"cps.html":                      {Data: []byte("<html><body>cps</body></html>")},
		"trusted_root.json":             {Data: []byte(`{"mediaType":"application/vnd.dev.sigstore.trustedroot+json;version=0.1"}`)},
		"signing-config.json":           {Data: []byte(`{"mediaType":"application/vnd.dev.sigstore.signingconfig+json;version=0.2"}`)},
		"certs/root-ca.crt":             {Data: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")},
		"certs/fulcio-ca.crt":           {Data: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")},
		"certs/spire-ca.crt":            {Data: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")},
		"certs/tsa.crt":                 {Data: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")},
		"keys/rekor-checkpoint.pub":     {Data: []byte("-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----\n")},
		"keys/tesseract-checkpoint.pub": {Data: []byte("-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----\n")},
	}
}

func TestNew_LoadsAllFiles(t *testing.T) {
	data := setupTestData(t)
	h := New(testLogger(t), data)

	expected := []string{
		"index.html", "favicon.svg", "cps.html",
		"trusted_root.json", "signing-config.json",
		"certs/root-ca.crt", "certs/fulcio-ca.crt", "certs/spire-ca.crt", "certs/tsa.crt",
		"keys/rekor-checkpoint.pub", "keys/tesseract-checkpoint.pub",
	}

	for _, name := range expected {
		if _, ok := h.files[name]; !ok {
			t.Errorf("file %q not loaded", name)
		}
		if len(h.files[name]) == 0 {
			t.Errorf("file %q is empty", name)
		}
	}
}

func TestNew_PanicsOnMissingFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on missing files, got nil")
		}
	}()

	New(testLogger(t), fstest.MapFS{})
}

func TestRoutes(t *testing.T) {
	data := setupTestData(t)
	lg := testLogger(t)
	h := New(lg, data)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	tests := []struct {
		path        string
		wantStatus  int
		wantType    string
		wantContain string
	}{
		{"/", http.StatusOK, "text/html", "index"},
		{"/favicon.svg", http.StatusOK, "image/svg+xml", "<svg>"},
		{"/cps", http.StatusOK, "text/html", "cps"},
		{"/.well-known/trusted_root.json", http.StatusOK, "application/json", "mediaType"},
		{"/.well-known/signing-config.json", http.StatusOK, "application/json", "mediaType"},
		{"/certs/root-ca.crt", http.StatusOK, "application/x-pem-file", "BEGIN CERTIFICATE"},
		{"/certs/fulcio-ca.crt", http.StatusOK, "application/x-pem-file", "BEGIN CERTIFICATE"},
		{"/certs/spire-ca.crt", http.StatusOK, "application/x-pem-file", "BEGIN CERTIFICATE"},
		{"/certs/tsa.crt", http.StatusOK, "application/x-pem-file", "BEGIN CERTIFICATE"},
		{"/keys/rekor-checkpoint.pub", http.StatusOK, "application/x-pem-file", "BEGIN PUBLIC KEY"},
		{"/keys/tesseract-checkpoint.pub", http.StatusOK, "application/x-pem-file", "BEGIN PUBLIC KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			ct := w.Header().Get("Content-Type")
			if ct != tt.wantType {
				t.Errorf("Content-Type = %q, want %q", ct, tt.wantType)
			}

			body := w.Body.String()
			if tt.wantContain != "" && !strings.Contains(body, tt.wantContain) {
				t.Errorf("body does not contain %q, got %q", tt.wantContain, body)
			}
		})
	}
}

func TestRoutes_NotFound(t *testing.T) {
	data := setupTestData(t)
	h := New(testLogger(t), data)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/nonexistent", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 404 or 405", w.Code)
	}
}
