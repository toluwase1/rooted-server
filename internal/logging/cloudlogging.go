package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
)

type Client struct {
	project  string
	services []string
	hc       *http.Client
}

func NewClient(project string, services ...string) (*Client, error) {
	ctx := context.Background()
	_, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/logging.read")
	if err != nil {
		return nil, fmt.Errorf("find credentials: %w", err)
	}
	return &Client{
		project:  project,
		services: services,
		hc:       &http.Client{Timeout: 30 * time.Second},
	}, nil
}

type LogEntry struct {
	Timestamp   string         `json:"timestamp"`
	Severity    string         `json:"severity"`
	TextPayload string         `json:"textPayload,omitempty"`
	JSONPayload map[string]any `json:"jsonPayload,omitempty"`
	HTTPRequest *HTTPRequest   `json:"httpRequest,omitempty"`
	InsertID    string         `json:"insertId"`
	Resource    *Resource      `json:"resource,omitempty"`
}

type HTTPRequest struct {
	RequestMethod string `json:"requestMethod"`
	RequestURL    string `json:"requestUrl"`
	Status        int    `json:"status"`
	Latency       string `json:"latency"`
}

type Resource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type listEntriesRequest struct {
	ResourceNames []string `json:"resourceNames"`
	Filter        string   `json:"filter"`
	OrderBy       string   `json:"orderBy"`
	PageSize      int      `json:"pageSize"`
	PageToken     string   `json:"pageToken,omitempty"`
}

type listEntriesResponse struct {
	Entries       []json.RawMessage `json:"entries"`
	NextPageToken string            `json:"nextPageToken"`
}

type QueryParams struct {
	Service   string
	Severity  string
	Search    string
	StartTime time.Time
	EndTime   time.Time
	PageSize  int
	PageToken string
}

type LogResult struct {
	Entries       []ParsedEntry `json:"entries"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

type ParsedEntry struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Severity   string         `json:"severity"`
	Service    string         `json:"service,omitempty"`
	Message    string         `json:"message"`
	Method     string         `json:"method,omitempty"`
	Path       string         `json:"path,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	DurationMs int64          `json:"duration_ms,omitempty"`
	TelegramID int64          `json:"telegram_id,omitempty"`
	Error      string         `json:"error,omitempty"`
	RawPayload map[string]any `json:"raw_payload,omitempty"`
}

func (c *Client) QueryLogs(ctx context.Context, params QueryParams) (*LogResult, error) {
	if params.PageSize <= 0 || params.PageSize > 500 {
		params.PageSize = 100
	}
	if params.EndTime.IsZero() {
		params.EndTime = time.Now().UTC()
	}
	if params.StartTime.IsZero() {
		params.StartTime = params.EndTime.Add(-24 * time.Hour)
	}

	reqBody := listEntriesRequest{
		ResourceNames: []string{"projects/" + c.project},
		Filter:        c.buildFilter(params),
		OrderBy:       "timestamp desc",
		PageSize:      params.PageSize,
		PageToken:     params.PageToken,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://logging.googleapis.com/v2/entries:list", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloud logging request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cloud logging API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var listResp listEntriesResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	result := &LogResult{NextPageToken: listResp.NextPageToken}
	for _, raw := range listResp.Entries {
		var entry LogEntry
		json.Unmarshal(raw, &entry)
		result.Entries = append(result.Entries, c.parseEntry(entry))
	}

	return result, nil
}

type LogStats struct {
	TotalEntries    int            `json:"total_entries"`
	TotalErrors     int            `json:"total_errors"`
	ErrorRate       float64        `json:"error_rate"`
	AvgResponseMs   float64        `json:"avg_response_ms"`
	BySeverity      map[string]int `json:"by_severity"`
	TopErrors       []TopError     `json:"top_errors"`
	RequestsPerHour []HourlyCount  `json:"requests_per_hour"`
	ErrorsPerHour   []HourlyCount  `json:"errors_per_hour"`
	TopPaths        []PathStat     `json:"top_paths"`
	TopStatusCodes  map[int]int    `json:"top_status_codes"`
}

type TopError struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

type HourlyCount struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

type PathStat struct {
	Method string  `json:"method"`
	Path   string  `json:"path"`
	Count  int     `json:"count"`
	AvgMs  float64 `json:"avg_ms"`
	Errors int     `json:"errors"`
}

func (c *Client) GetStats(ctx context.Context, startTime, endTime time.Time) (*LogStats, error) {
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}
	if startTime.IsZero() {
		startTime = endTime.Add(-24 * time.Hour)
	}

	allEntries := []ParsedEntry{}
	pageToken := ""

	for i := 0; i < 5; i++ {
		result, err := c.QueryLogs(ctx, QueryParams{
			StartTime: startTime,
			EndTime:   endTime,
			PageSize:  200,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, result.Entries...)
		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return c.aggregateStats(allEntries, startTime, endTime), nil
}

func (c *Client) aggregateStats(entries []ParsedEntry, startTime, endTime time.Time) *LogStats {
	stats := &LogStats{
		BySeverity:     make(map[string]int),
		TopStatusCodes: make(map[int]int),
	}

	errorCounts := map[string]int{}
	pathStats := map[string]*PathStat{}
	hourlyReqs := map[string]int{}
	hourlyErrs := map[string]int{}
	var totalMs int64
	var msCount int

	for _, e := range entries {
		stats.TotalEntries++
		stats.BySeverity[e.Severity]++

		isError := e.Severity == "ERROR" || e.StatusCode >= 500
		if isError {
			stats.TotalErrors++
		}

		if e.StatusCode > 0 {
			stats.TopStatusCodes[e.StatusCode]++
		}
		if e.DurationMs > 0 {
			totalMs += e.DurationMs
			msCount++
		}

		hour := e.Timestamp.Truncate(time.Hour).Format("2006-01-02T15:00")
		hourlyReqs[hour]++
		if isError {
			hourlyErrs[hour]++
		}

		if isError && e.Message != "" {
			errorCounts[e.Message]++
		}

		if e.Method != "" && e.Path != "" {
			key := e.Method + " " + e.Path
			if _, ok := pathStats[key]; !ok {
				pathStats[key] = &PathStat{Method: e.Method, Path: e.Path}
			}
			ps := pathStats[key]
			ps.Count++
			if e.DurationMs > 0 {
				ps.AvgMs += float64(e.DurationMs)
			}
			if isError {
				ps.Errors++
			}
		}
	}

	if stats.TotalEntries > 0 {
		stats.ErrorRate = float64(stats.TotalErrors) / float64(stats.TotalEntries) * 100
	}
	if msCount > 0 {
		stats.AvgResponseMs = float64(totalMs) / float64(msCount)
	}

	for msg, count := range errorCounts {
		stats.TopErrors = append(stats.TopErrors, TopError{Message: msg, Count: count})
	}
	sortTopErrors(stats.TopErrors)
	if len(stats.TopErrors) > 10 {
		stats.TopErrors = stats.TopErrors[:10]
	}

	for _, ps := range pathStats {
		if ps.Count > 0 && ps.AvgMs > 0 {
			ps.AvgMs /= float64(ps.Count)
		}
		stats.TopPaths = append(stats.TopPaths, *ps)
	}
	sortPathStats(stats.TopPaths)
	if len(stats.TopPaths) > 15 {
		stats.TopPaths = stats.TopPaths[:15]
	}

	start := startTime.Truncate(time.Hour)
	end := endTime.Truncate(time.Hour)
	for h := start; !h.After(end); h = h.Add(time.Hour) {
		key := h.Format("2006-01-02T15:00")
		stats.RequestsPerHour = append(stats.RequestsPerHour, HourlyCount{Hour: key, Count: hourlyReqs[key]})
		stats.ErrorsPerHour = append(stats.ErrorsPerHour, HourlyCount{Hour: key, Count: hourlyErrs[key]})
	}

	return stats
}

func (c *Client) buildFilter(params QueryParams) string {
	var serviceParts []string
	if params.Service != "" {
		serviceParts = append(serviceParts, fmt.Sprintf(`resource.labels.service_name="%s"`, params.Service))
	} else {
		for _, s := range c.services {
			serviceParts = append(serviceParts, fmt.Sprintf(`resource.labels.service_name="%s"`, s))
		}
	}
	serviceFilter := "(" + strings.Join(serviceParts, " OR ") + ")"

	parts := []string{
		`resource.type="cloud_run_revision"`,
		serviceFilter,
		fmt.Sprintf(`timestamp>="%s"`, params.StartTime.Format(time.RFC3339)),
		fmt.Sprintf(`timestamp<="%s"`, params.EndTime.Format(time.RFC3339)),
	}

	if params.Severity != "" {
		parts = append(parts, fmt.Sprintf(`severity>=%s`, strings.ToUpper(params.Severity)))
	}
	if params.Search != "" {
		parts = append(parts, fmt.Sprintf(`textPayload:"%s" OR jsonPayload.path:"%s" OR jsonPayload.error:"%s"`,
			params.Search, params.Search, params.Search))
	}

	return strings.Join(parts, "\n")
}

func (c *Client) parseEntry(entry LogEntry) ParsedEntry {
	parsed := ParsedEntry{
		ID:       entry.InsertID,
		Severity: entry.Severity,
	}

	if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
		parsed.Timestamp = t
	}

	if entry.JSONPayload != nil {
		parsed.RawPayload = entry.JSONPayload
		if msg, ok := entry.JSONPayload["msg"].(string); ok {
			parsed.Message = msg
		}
		if method, ok := entry.JSONPayload["method"].(string); ok {
			parsed.Method = method
		}
		if path, ok := entry.JSONPayload["path"].(string); ok {
			parsed.Path = path
		}
		if status, ok := entry.JSONPayload["status"].(float64); ok {
			parsed.StatusCode = int(status)
		}
		if dur, ok := entry.JSONPayload["latency_ms"].(float64); ok {
			parsed.DurationMs = int64(dur)
		}
		if tgID, ok := entry.JSONPayload["telegram_id"].(float64); ok {
			parsed.TelegramID = int64(tgID)
		}
		if errMsg, ok := entry.JSONPayload["response_body"].(string); ok {
			parsed.Error = errMsg
		}
	} else if entry.TextPayload != "" {
		parsed.Message = entry.TextPayload
	}

	if entry.HTTPRequest != nil {
		if parsed.Method == "" {
			parsed.Method = entry.HTTPRequest.RequestMethod
		}
		if parsed.Path == "" {
			parsed.Path = entry.HTTPRequest.RequestURL
		}
		if parsed.StatusCode == 0 {
			parsed.StatusCode = entry.HTTPRequest.Status
		}
	}

	if entry.Resource != nil && entry.Resource.Labels != nil {
		if svc, ok := entry.Resource.Labels["service_name"]; ok {
			parsed.Service = svc
		}
	}

	if parsed.Severity == "" {
		parsed.Severity = "DEFAULT"
	}

	return parsed
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/logging.read")
	if err != nil {
		return "", err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func sortTopErrors(errs []TopError) {
	for i := range errs {
		for j := i + 1; j < len(errs); j++ {
			if errs[j].Count > errs[i].Count {
				errs[i], errs[j] = errs[j], errs[i]
			}
		}
	}
}

func sortPathStats(paths []PathStat) {
	for i := range paths {
		for j := i + 1; j < len(paths); j++ {
			if paths[j].Count > paths[i].Count {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}
}
