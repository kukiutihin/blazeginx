package httpserver

import (
	"blazeginx/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func writeStaticFile(t *testing.T, root string, name string, body string) {
	t.Helper()

	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create static dir: %s", err)
	}

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("failed to write static file: %s", err)
	}
}

func serverWithStatic(static config.Static) string {
	server := NewServer(
		nil,
		[]handler{
			{
				"/*",
				NewStaticHandler(static),
			},
		},
	)

	return server.URL
}

func TestStaticFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeStaticFile(t, root, "assets/app.js", "console.log('ok')")

	url := serverWithStatic(config.Static{
		Enabled: true,
		Root:    root,
	})

	status, body := DoGet(url+"/assets/app.js", t)
	if status != http.StatusOK {
		t.Fatalf("Expected status: %d, but got: %d", http.StatusOK, status)
	}
	if body != "console.log('ok')" {
		t.Fatalf("Expected body: %s, but got: %s", "console.log('ok')", body)
	}
}

func TestStaticFallbackToIndex(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeStaticFile(t, root, "index.html", "<html>spa</html>")

	url := serverWithStatic(config.Static{
		Enabled: true,
		Root:    root,
	})

	status, body := DoGet(url+"/dashboard/settings", t)
	if status != http.StatusOK {
		t.Fatalf("Expected status: %d, but got: %d", http.StatusOK, status)
	}
	if body != "<html>spa</html>" {
		t.Fatalf("Expected body: %s, but got: %s", "<html>spa</html>", body)
	}
}

func TestStaticMissingAsset(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeStaticFile(t, root, "index.html", "<html>spa</html>")

	url := serverWithStatic(config.Static{
		Enabled: true,
		Root:    root,
	})

	status := DoGetCode(url+"/assets/missing.js", t)
	if status != http.StatusNotFound {
		t.Fatalf("Expected status: %d, but got: %d", http.StatusNotFound, status)
	}
}

func TestStaticDisabled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeStaticFile(t, root, "index.html", "<html>spa</html>")

	url := serverWithStatic(config.Static{
		Enabled: false,
		Root:    root,
	})

	status := DoGetCode(url+"/dashboard", t)
	if status != http.StatusNotFound {
		t.Fatalf("Expected status: %d, but got: %d", http.StatusNotFound, status)
	}
}
