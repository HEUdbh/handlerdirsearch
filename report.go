package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func appendMarkdownReport(reportPath, inputFilePath string, response ScanResponse) error {
	file, err := os.OpenFile(reportPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open report file: %w", err)
	}
	defer file.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("## Scan Report - %s\n", now))
	builder.WriteString(fmt.Sprintf("- Input File: `%s`\n", inputFilePath))
	builder.WriteString(fmt.Sprintf("- Total 200 Lines: %d\n", response.Total200Lines))
	builder.WriteString(fmt.Sprintf("- Total URLs: %d\n", response.TotalURLs))
	builder.WriteString(fmt.Sprintf("- Succeeded: %d\n", response.Succeeded))
	builder.WriteString(fmt.Sprintf("- Failed: %d\n\n", response.Failed))
	builder.WriteString("| URL | Title | Components | Error |\n")
	builder.WriteString("| --- | --- | --- | --- |\n")

	if len(response.Rows) == 0 {
		builder.WriteString("| N/A | N/A | N/A | No URL found from 200 lines |\n\n")
	} else {
		for _, row := range response.Rows {
			components := "N/A"
			if len(row.Components) > 0 {
				components = strings.Join(row.Components, ", ")
			}

			errorText := row.Error
			if errorText == "" {
				errorText = "-"
			}

			builder.WriteString("| ")
			builder.WriteString(escapeMarkdownCell(row.URL))
			builder.WriteString(" | ")
			builder.WriteString(escapeMarkdownCell(row.Title))
			builder.WriteString(" | ")
			builder.WriteString(escapeMarkdownCell(components))
			builder.WriteString(" | ")
			builder.WriteString(escapeMarkdownCell(errorText))
			builder.WriteString(" |\n")
		}
		builder.WriteString("\n")
	}

	if _, err := file.WriteString(builder.String()); err != nil {
		return fmt.Errorf("write report file: %w", err)
	}

	return nil
}

func escapeMarkdownCell(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", "<br/>")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
