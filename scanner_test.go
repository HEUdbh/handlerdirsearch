package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInputFileDirsearchStyle(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.txt")

	content := strings.Join([]string{
		"200 123B 0.001s http://example.com/a",
		"301 123B 0.001s http://example.com/b",
		"403 123B 0.001s https://example.org/blocked",
		"404 123B 0.001s http://example.com/ignore",
		"  200 88B 0.003s https://example.org/login?x=1",
		"403 no-url-here",
		"301 duplicate http://example.com/b",
	}, "\n")

	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	urls, totalMatched, err := parseInputFile(inputPath)
	if err != nil {
		t.Fatalf("parse input file: %v", err)
	}

	if totalMatched != 6 {
		t.Fatalf("expected total matched lines=6, got %d", totalMatched)
	}

	expected := []string{
		"http://example.com/a",
		"http://example.com/b",
		"https://example.org/blocked",
		"https://example.org/login?x=1",
	}
	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %#v", len(expected), len(urls), urls)
	}

	for i := range expected {
		if urls[i] != expected[i] {
			t.Fatalf("expected url[%d]=%s, got %s", i, expected[i], urls[i])
		}
	}
}

func TestRunScanWorkersExtractsSignalsAndMarksFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("X-Powered-By", "PHP/8.2")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!doctype html><html><head><title>Test Site</title><meta name="generator" content="WordPress 6.0"></head><body>wp-content next.js</body></html>`))
		case "/bad":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	rows := runScanWorkers([]string{server.URL + "/ok", server.URL + "/bad"}, ScanRequest{
		Concurrency:    2,
		TimeoutSeconds: 5,
		FollowRedirect: true,
	})

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0].Title != "Test Site" {
		t.Fatalf("expected title Test Site, got %q", rows[0].Title)
	}

	components := strings.Join(rows[0].Components, " | ")
	for _, expected := range []string{"X-Powered-By: PHP/8.2", "Meta Generator: WordPress 6.0", "WordPress", "Next.js"} {
		if !strings.Contains(components, expected) {
			t.Fatalf("expected component %q in %q", expected, components)
		}
	}

	if rows[0].Error != "" {
		t.Fatalf("expected no error for ok row, got %q", rows[0].Error)
	}

	if rows[1].Error == "" {
		t.Fatalf("expected error for bad row")
	}
	if !strings.Contains(rows[1].Error, "HTTP 500") {
		t.Fatalf("expected HTTP 500 in error, got %q", rows[1].Error)
	}
}
