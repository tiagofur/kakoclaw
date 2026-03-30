package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
)

type AnalyticsSummary struct {
	TotalPosts       int                      `json:"total_posts"`
	TotalImpressions int64                    `json:"total_impressions"`
	TotalEngagements int64                    `json:"total_engagements"`
	EngagementRate   float64                  `json:"engagement_rate"`
	ByPlatform       map[string]PlatformStats `json:"by_platform"`
	TrendDates       []string                 `json:"trend_dates"`
	TrendImpressions []int64                  `json:"trend_impressions"`
	HasLegacyReports bool                     `json:"has_legacy_reports"`
}

type PlatformStats struct {
	Posts       int     `json:"posts"`
	Impressions int64   `json:"impressions"`
	Engagements int64   `json:"engagements"`
	Rate        float64 `json:"rate"`
}

var marketingAnalyticsFilePattern = regexp.MustCompile(`^([a-z0-9-]+)_(.+)_([0-9]{4}-[0-9]{2}-[0-9]{2})\.json$`)

func (s *Server) handleMarketingAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	accountName, campaignName, err := parseMarketingAnalyticsSummaryPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	analyticsDir := filepath.Join(userWorkspace, "marketing", accountName, campaignName, "analytics")
	summary := AnalyticsSummary{
		ByPlatform:       map[string]PlatformStats{},
		TrendDates:       []string{},
		TrendImpressions: []int64{},
	}
	if _, err := os.Stat(analyticsDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, summary)
		return
	}

	entries, err := os.ReadDir(analyticsDir)
	if err != nil {
		http.Error(w, "failed to read analytics directory", http.StatusInternalServerError)
		return
	}

	trend := map[string]int64{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.ToLower(filepath.Ext(name)) != ".json" {
			summary.HasLegacyReports = true
			continue
		}
		match := marketingAnalyticsFilePattern.FindStringSubmatch(name)
		if len(match) != 4 {
			continue
		}
		platform := match[1]
		date := match[3]

		data, err := os.ReadFile(filepath.Join(analyticsDir, name))
		if err != nil {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			continue
		}

		impressions := analyticsMetricInt64(payload, "impressions")
		engagements := analyticsMetricInt64(payload, "engagements")
		if engagements == 0 {
			engagements = analyticsMetricInt64(payload, "likes") + analyticsMetricInt64(payload, "comments") + analyticsMetricInt64(payload, "shares") + analyticsMetricInt64(payload, "clicks")
		}

		summary.TotalPosts++
		summary.TotalImpressions += impressions
		summary.TotalEngagements += engagements

		stats := summary.ByPlatform[platform]
		stats.Posts++
		stats.Impressions += impressions
		stats.Engagements += engagements
		summary.ByPlatform[platform] = stats

		trend[date] += impressions
	}

	if summary.TotalImpressions > 0 {
		summary.EngagementRate = float64(summary.TotalEngagements) / float64(summary.TotalImpressions)
	}
	for platform, stats := range summary.ByPlatform {
		if stats.Impressions > 0 {
			stats.Rate = float64(stats.Engagements) / float64(stats.Impressions)
		}
		summary.ByPlatform[platform] = stats
	}

	dates := make([]string, 0, len(trend))
	for date := range trend {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		summary.TrendDates = append(summary.TrendDates, date)
		summary.TrendImpressions = append(summary.TrendImpressions, trend[date])
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, summary)
}

func parseMarketingAnalyticsSummaryPath(path string) (string, string, error) {
	suffix := strings.TrimPrefix(path, "/api/v1/marketing/campaigns/")
	parts := strings.Split(suffix, "/")
	if len(parts) != 4 || parts[2] != "analytics" || parts[3] != "summary" {
		return "", "", httpErrorf("invalid campaign path")
	}
	if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", httpErrorf("invalid campaign path")
	}
	return parts[0], parts[1], nil
}

func analyticsMetricInt64(payload map[string]any, key string) int64 {
	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}
