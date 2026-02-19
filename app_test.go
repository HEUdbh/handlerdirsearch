package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunScanCreatesMarkdownReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Powered-By", "PHP/8.2")
		_, _ = w.Write([]byte("<html><head><title>Demo</title></head><body>wp-content</body></html>"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "source.txt")
	input := "200 120B 0.001s " + server.URL + "/home\n"
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	app := NewApp()
	result, err := app.RunScan(ScanRequest{
		InputFilePath:  inputPath,
		Concurrency:    30,
		TimeoutSeconds: 5,
		FollowRedirect: true,
	})
	if err != nil {
		t.Fatalf("run scan: %v", err)
	}

	if result.TotalURLs != 1 || result.Succeeded != 1 || result.Failed != 0 {
		t.Fatalf("unexpected stats: %+v", result)
	}

	reportPath := filepath.Join(tempDir, "scan_report.md")
	if result.ReportPath != reportPath {
		t.Fatalf("expected report path %s, got %s", reportPath, result.ReportPath)
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	report := string(reportBytes)
	if !strings.Contains(report, "| URL | Title | Components | Error |") {
		t.Fatalf("report table header missing: %s", report)
	}
	if !strings.Contains(report, server.URL+"/home") {
		t.Fatalf("report URL missing: %s", report)
	}
	if !strings.Contains(report, "Demo") {
		t.Fatalf("report title missing: %s", report)
	}
}
