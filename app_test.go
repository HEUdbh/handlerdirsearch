package main

import (
	"errors"
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

	reportPath := filepath.Join(tempDir, "source_report.md")
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

func TestRunScanWritesReportToOutputDir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><head><title>Demo</title></head><body>ok</body></html>"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "source.txt")
	input := "301 120B 0.001s " + server.URL + "/home\n"
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	outputDir := filepath.Join(tempDir, "reports")
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	app := NewApp()
	result, err := app.RunScan(ScanRequest{
		InputFilePath:  inputPath,
		OutputDir:      outputDir,
		Concurrency:    30,
		TimeoutSeconds: 5,
		FollowRedirect: true,
	})
	if err != nil {
		t.Fatalf("run scan: %v", err)
	}

	expectedPath := filepath.Join(outputDir, "source_report.md")
	if result.ReportPath != expectedPath {
		t.Fatalf("expected report path %s, got %s", expectedPath, result.ReportPath)
	}
	if filepath.Ext(result.ReportPath) != ".md" {
		t.Fatalf("expected markdown report extension, got %s", filepath.Ext(result.ReportPath))
	}
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected report file at %s: %v", expectedPath, err)
	}
}

func TestRunScanDeletesInputFileWhenRequested(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><head><title>Demo</title></head><body>ok</body></html>"))
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
		InputFilePath:        inputPath,
		Concurrency:          30,
		TimeoutSeconds:       5,
		FollowRedirect:       true,
		DeleteSourceAfterRun: true,
	})
	if err != nil {
		t.Fatalf("run scan: %v", err)
	}

	reportPath := filepath.Join(tempDir, "source_report.md")
	if result.ReportPath != reportPath {
		t.Fatalf("expected report path %s, got %s", reportPath, result.ReportPath)
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("expected report file at %s: %v", reportPath, err)
	}
	if _, err := os.Stat(inputPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected input file to be deleted, got err=%v", err)
	}
}

func TestRunScanDoesNotDeleteInputFileByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><head><title>Demo</title></head><body>ok</body></html>"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "source.txt")
	input := "200 120B 0.001s " + server.URL + "/home\n"
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	app := NewApp()
	_, err := app.RunScan(ScanRequest{
		InputFilePath:  inputPath,
		Concurrency:    30,
		TimeoutSeconds: 5,
		FollowRedirect: true,
	})
	if err != nil {
		t.Fatalf("run scan: %v", err)
	}

	if _, err := os.Stat(inputPath); err != nil {
		t.Fatalf("expected input file to remain, stat err=%v", err)
	}
}

func TestRunScanReturnsErrorWhenDeleteFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><head><title>Demo</title></head><body>ok</body></html>"))
	}))
	defer server.Close()

	originalRemoveInputFile := removeInputFile
	removeInputFile = func(string) error {
		return errors.New("mock delete failure")
	}
	t.Cleanup(func() {
		removeInputFile = originalRemoveInputFile
	})

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "source.txt")
	input := "200 120B 0.001s " + server.URL + "/home\n"
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	app := NewApp()
	_, err := app.RunScan(ScanRequest{
		InputFilePath:        inputPath,
		Concurrency:          30,
		TimeoutSeconds:       5,
		FollowRedirect:       true,
		DeleteSourceAfterRun: true,
	})
	if err == nil {
		t.Fatalf("expected delete failure error")
	}
	if !strings.Contains(err.Error(), "删除源文件失败") {
		t.Fatalf("expected delete failure message, got %v", err)
	}

	reportPath := filepath.Join(tempDir, "source_report.md")
	if _, statErr := os.Stat(reportPath); statErr != nil {
		t.Fatalf("expected report to be written before delete failure: %v", statErr)
	}
	if _, statErr := os.Stat(inputPath); statErr != nil {
		t.Fatalf("expected input file to remain after delete failure: %v", statErr)
	}
}
