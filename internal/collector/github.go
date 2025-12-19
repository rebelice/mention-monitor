package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rebelice/mention-monitor/internal/models"
)

// GitHub collects mentions from GitHub Issues and Discussions
type GitHub struct {
	Token string
}

type ghSearchResponse struct {
	Items []ghItem `json:"items"`
}

type ghItem struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	User      ghUser    `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

type ghUser struct {
	Login string `json:"login"`
}

func (g *GitHub) Name() string { return "github" }

func (g *GitHub) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		// Search issues
		issues, err := g.searchIssues(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, issues...)

		// Search code (for package imports)
		code, err := g.searchCode(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, code...)
	}

	return mentions, nil
}

func (g *GitHub) searchIssues(ctx context.Context, keyword string) ([]models.Mention, error) {
	// Search issues created in last 24 hours
	since := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	query := fmt.Sprintf("%s created:>%s", keyword, since)
	apiURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=created&order=desc", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	var result ghSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range result.Items {
		mentions = append(mentions, models.Mention{
			ID:           fmt.Sprintf("github_%d", item.ID),
			Source:       "github",
			Type:         "issue",
			Keyword:      keyword,
			Title:        item.Title,
			Content:      truncate(item.Body, 500),
			URL:          item.HTMLURL,
			Author:       item.User.Login,
			DiscoveredAt: time.Now().UTC(),
			PublishedAt:  item.CreatedAt,
		})
	}

	return mentions, nil
}

type ghCodeSearchResponse struct {
	Items []ghCodeItem `json:"items"`
}

type ghCodeItem struct {
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	HTMLURL    string     `json:"html_url"`
	Repository ghCodeRepo `json:"repository"`
}

type ghCodeRepo struct {
	FullName string `json:"full_name"`
}

func (g *GitHub) searchCode(ctx context.Context, keyword string) ([]models.Mention, error) {
	// Search for package imports (limited to go.mod files)
	query := fmt.Sprintf("%s filename:go.mod", keyword)
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&sort=indexed&order=desc&per_page=10", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "token "+g.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github code search returned status %d", resp.StatusCode)
	}

	var result ghCodeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range result.Items {
		mentions = append(mentions, models.Mention{
			ID:           fmt.Sprintf("github_code_%s_%s", item.Repository.FullName, item.Path),
			Source:       "github",
			Type:         "code",
			Keyword:      keyword,
			Title:        fmt.Sprintf("Used in %s", item.Repository.FullName),
			Content:      fmt.Sprintf("Found in %s", item.Path),
			URL:          item.HTMLURL,
			Author:       item.Repository.FullName,
			DiscoveredAt: time.Now().UTC(),
		})
	}

	return mentions, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
