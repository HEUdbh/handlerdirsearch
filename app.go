package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	defaultConcurrency   = 30
	defaultTimeoutSecond = 5
)

// App struct
type App struct {
	ctx context.Context
}

type ScanRequest struct {
	InputFilePath  string `json:"inputFilePath"`
	Concurrency    int    `json:"concurrency"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	FollowRedirect bool   `json:"followRedirect"`
}

type ScanRow struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Components []string `json:"components"`
	Error      string   `json:"error"`
}

type ScanResponse struct {
	ReportPath    string    `json:"reportPath"`
	Total200Lines int       `json:"total200Lines"`
	TotalURLs     int       `json:"totalUrls"`
	Succeeded     int       `json:"succeeded"`
	Failed        int       `json:"failed"`
	Rows          []ScanRow `json:"rows"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SelectInputFile() (string, error) {
	if a.ctx == nil {
		return "", errors.New("\u5e94\u7528\u4e0a\u4e0b\u6587\u5c1a\u672a\u521d\u59cb\u5316")
	}

	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "\u9009\u62e9 URL \u6e90\u6587\u672c\u6587\u4ef6",
		Filters: []runtime.FileFilter{
			{DisplayName: "\u6587\u672c\u6587\u4ef6", Pattern: "*.txt;*.log;*.md;*.csv"},
			{DisplayName: "\u6240\u6709\u6587\u4ef6", Pattern: "*.*"},
		},
	})
}

func (a *App) RunScan(request ScanRequest) (ScanResponse, error) {
	request = normalizeScanRequest(request)

	if request.InputFilePath == "" {
		return ScanResponse{}, errors.New("\u8bf7\u8f93\u5165\u8f93\u5165\u6587\u4ef6\u8def\u5f84")
	}

	urls, total200Lines, err := parseInputFile(request.InputFilePath)
	if err != nil {
		return ScanResponse{}, err
	}

	response := ScanResponse{
		Total200Lines: total200Lines,
		TotalURLs:     len(urls),
	}

	if len(urls) > 0 {
		response.Rows = runScanWorkers(urls, request)
		for _, row := range response.Rows {
			if row.Error == "" {
				response.Succeeded++
			} else {
				response.Failed++
			}
		}
	}

	reportPath := filepath.Join(filepath.Dir(request.InputFilePath), "scan_report.md")
	if err := appendMarkdownReport(reportPath, request.InputFilePath, response); err != nil {
		return ScanResponse{}, err
	}

	response.ReportPath = reportPath
	return response, nil
}

func normalizeScanRequest(request ScanRequest) ScanRequest {
	if request.Concurrency <= 0 {
		request.Concurrency = defaultConcurrency
	}

	if request.TimeoutSeconds <= 0 {
		request.TimeoutSeconds = defaultTimeoutSecond
	}

	if request.Concurrency > 100 {
		request.Concurrency = 100
	}

	if request.TimeoutSeconds > 120 {
		request.TimeoutSeconds = 120
	}

	return request
}
