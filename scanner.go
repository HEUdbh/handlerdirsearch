package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	maxBodySize      = 2 << 20
)

var (
	statusMatchRegex = regexp.MustCompile(`^\s*(200|301|403)\b`)
	urlRegex         = regexp.MustCompile(`https?://[^\s"'<>]+`)
)

type indexedURL struct {
	Index int
	URL   string
}

type indexedRow struct {
	Index int
	Row   ScanRow
}

func parseInputFile(path string) ([]string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	seen := make(map[string]struct{})
	urls := make([]string, 0)
	totalMatchedLines := 0

	for scanner.Scan() {
		line := scanner.Text()
		if !statusMatchRegex.MatchString(line) {
			continue
		}

		totalMatchedLines++
		matched := urlRegex.FindString(line)
		if matched == "" {
			continue
		}

		if _, ok := seen[matched]; ok {
			continue
		}

		seen[matched] = struct{}{}
		urls = append(urls, matched)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("read input file: %w", err)
	}

	return urls, totalMatchedLines, nil
}

func runScanWorkers(urls []string, request ScanRequest) []ScanRow {
	if len(urls) == 0 {
		return nil
	}

	concurrency := request.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	if concurrency > len(urls) {
		concurrency = len(urls)
	}

	results := make([]ScanRow, len(urls))
	jobs := make(chan indexedURL)
	out := make(chan indexedRow)

	client := newHTTPClient(request)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				out <- indexedRow{Index: job.Index, Row: scanURL(client, job.URL)}
			}
		}()
	}

	go func() {
		for i, url := range urls {
			jobs <- indexedURL{Index: i, URL: url}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	for item := range out {
		results[item.Index] = item.Row
	}

	return results
}

func newHTTPClient(request ScanRequest) *http.Client {
	timeoutSeconds := request.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeoutSecond
	}

	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
	if !request.FollowRedirect {
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func scanURL(client *http.Client, targetURL string) ScanRow {
	row := ScanRow{
		URL:        targetURL,
		Title:      "N/A",
		Components: []string{"N/A"},
	}

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if readErr != nil {
		row.Error = readErr.Error()
	}

	title, generator := extractHTMLSignals(body)
	if title != "" {
		row.Title = title
	}
	row.Components = extractComponents(resp, body, generator)

	if resp.StatusCode >= http.StatusBadRequest {
		if row.Error == "" {
			row.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		} else {
			row.Error = fmt.Sprintf("HTTP %d; %s", resp.StatusCode, row.Error)
		}
	}

	return row
}

func extractHTMLSignals(body []byte) (string, string) {
	if len(body) == 0 {
		return "", ""
	}

	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	inTitle := false
	titleBuilder := strings.Builder{}
	generator := ""

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return strings.TrimSpace(html.UnescapeString(titleBuilder.String())), strings.TrimSpace(html.UnescapeString(generator))
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if strings.EqualFold(token.Data, "title") {
				inTitle = true
				continue
			}
			if !strings.EqualFold(token.Data, "meta") {
				continue
			}

			name := ""
			content := ""
			for _, attr := range token.Attr {
				if strings.EqualFold(attr.Key, "name") {
					name = strings.TrimSpace(attr.Val)
				}
				if strings.EqualFold(attr.Key, "content") {
					content = strings.TrimSpace(attr.Val)
				}
			}
			if strings.EqualFold(name, "generator") && content != "" {
				generator = content
			}
		case html.TextToken:
			if inTitle {
				titleBuilder.WriteString(tokenizer.Token().Data)
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			if strings.EqualFold(token.Data, "title") {
				inTitle = false
			}
		}
	}
}

func extractComponents(resp *http.Response, body []byte, generator string) []string {
	components := make([]string, 0)
	seen := make(map[string]struct{})

	for _, key := range []string{"Server", "X-Powered-By", "Via", "X-AspNet-Version", "X-AspNetMvc-Version"} {
		value := strings.TrimSpace(resp.Header.Get(key))
		if value == "" {
			continue
		}
		addUniqueComponent(&components, seen, fmt.Sprintf("%s: %s", key, value))
	}

	if generator != "" {
		addUniqueComponent(&components, seen, fmt.Sprintf("Meta Generator: %s", generator))
	}

	lowerBody := strings.ToLower(string(body))
	detectBodyComponents(lowerBody, &components, seen)

	if len(components) == 0 {
		return []string{"N/A"}
	}

	return components
}

func detectBodyComponents(body string, components *[]string, seen map[string]struct{}) {
	type marker struct {
		contains []string
		name     string
	}

	markers := []marker{
		{contains: []string{"wp-content", "wordpress"}, name: "WordPress"},
		{contains: []string{"drupal-settings-json", "drupal"}, name: "Drupal"},
		{contains: []string{"content=\"joomla", "joomla!"}, name: "Joomla"},
		{contains: []string{"__next", "next.js"}, name: "Next.js"},
		{contains: []string{"__nuxt", "nuxt"}, name: "Nuxt"},
		{contains: []string{"reactroot", "data-reactroot", "react-dom"}, name: "React"},
		{contains: []string{"data-v-", "vue.js", "vue.runtime"}, name: "Vue"},
		{contains: []string{"__viewstate", "asp.net"}, name: "ASP.NET"},
		{contains: []string{".php", "<?php", "php/"}, name: "PHP"},
		{contains: []string{"jsessionid", "java servlet", "jsp"}, name: "Java"},
	}

	for _, marker := range markers {
		for _, token := range marker.contains {
			if strings.Contains(body, token) {
				addUniqueComponent(components, seen, marker.name)
				break
			}
		}
	}
}

func addUniqueComponent(components *[]string, seen map[string]struct{}, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}

	key := strings.ToLower(trimmed)
	if _, ok := seen[key]; ok {
		return
	}

	seen[key] = struct{}{}
	*components = append(*components, trimmed)
}
