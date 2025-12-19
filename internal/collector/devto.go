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

// DevTo collects mentions from Dev.to via API
type DevTo struct{}

type devtoArticle struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	User        devtoUser `json:"user"`
	CreatedAt   string    `json:"created_at"`
}

type devtoUser struct {
	Username string `json:"username"`
}

func (d *DevTo) Name() string { return "devto" }

func (d *DevTo) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		results, err := d.search(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (d *DevTo) search(ctx context.Context, keyword string) ([]models.Mention, error) {
	apiURL := fmt.Sprintf("https://dev.to/api/articles?tag=%s&per_page=30", url.QueryEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("devto returned status %d", resp.StatusCode)
	}

	var articles []devtoArticle
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, a := range articles {
		// Filter by keyword in title or description
		text := a.Title + " " + a.Description
		if found, matchedKw := ContainsKeyword(text, []string{keyword}); !found {
			continue
		} else {
			m := models.Mention{
				ID:           fmt.Sprintf("devto_%d", a.ID),
				Source:       "devto",
				Type:         "article",
				Keyword:      matchedKw,
				Title:        a.Title,
				Content:      a.Description,
				URL:          a.URL,
				Author:       a.User.Username,
				DiscoveredAt: time.Now().UTC(),
			}

			if t, err := time.Parse(time.RFC3339, a.CreatedAt); err == nil {
				m.PublishedAt = t
			}

			mentions = append(mentions, m)
		}
	}

	return mentions, nil
}
