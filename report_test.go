package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendMarkdownReportAppendsSectionsAndEscapesCells(t *testing.T) {
	tempDir := t.TempDir()
	reportPath := filepath.Join(tempDir, "scan_report.md")

	response := ScanResponse{
		Total200Lines: 1,
		TotalURLs:     1,
		Succeeded:     1,
		Failed:        0,
		Rows: []ScanRow{
			{
				URL:        "https://example.com/a|b",
				Title:      "Line1\nLine2",
				Components: []string{"Server: nginx|1.25", "WordPress"},
				Error:      "",
			},
		},
	}

	if err := appendMarkdownReport(reportPath, filepath.Join(tempDir, "input.txt"), response); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := appendMarkdownReport(reportPath, filepath.Join(tempDir, "input.txt"), response); err != nil {
		t.Fatalf("second append: %v", err)
	}

	contentBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	content := string(contentBytes)

	if strings.Count(content, "## Scan Report - ") != 2 {
		t.Fatalf("expected 2 report sections, got content: %s", content)
	}

	if !strings.Contains(content, "https://example.com/a\\|b") {
		t.Fatalf("expected escaped URL pipe in report: %s", content)
	}
	if !strings.Contains(content, "Line1<br/>Line2") {
		t.Fatalf("expected newline conversion in report: %s", content)
	}
	if !strings.Contains(content, "Server: nginx\\|1.25") {
		t.Fatalf("expected escaped component pipe in report: %s", content)
	}
}
