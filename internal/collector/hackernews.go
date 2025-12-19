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

// HackerNews collects mentions from Hacker News via Algolia API
type HackerNews struct{}

type hnSearchResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID    string   `json:"objectID"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Author      string   `json:"author"`
	StoryText   string   `json:"story_text"`
	CommentText string   `json:"comment_text"`
	StoryTitle  string   `json:"story_title"`
	StoryURL    string   `json:"story_url"`
	ParentID    int      `json:"parent_id"`
	CreatedAt   string   `json:"created_at"`
	Tags        []string `json:"_tags"`
}

func (h *HackerNews) Name() string { return "hackernews" }

func (h *HackerNews) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		// Search stories
		stories, err := h.search(ctx, kw, "story")
		if err != nil {
			continue
		}
		mentions = append(mentions, stories...)

		// Search comments
		comments, err := h.search(ctx, kw, "comment")
		if err != nil {
			continue
		}
		mentions = append(mentions, comments...)
	}

	return mentions, nil
}

func (h *HackerNews) search(ctx context.Context, keyword, tag string) ([]models.Mention, error) {
	// Search last 24 hours
	apiURL := fmt.Sprintf(
		"https://hn.algolia.com/api/v1/search_by_date?query=%s&tags=%s&numericFilters=created_at_i>%d",
		url.QueryEscape(keyword),
		tag,
		time.Now().Add(-24*time.Hour).Unix(),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result hnSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, hit := range result.Hits {
		m := models.Mention{
			ID:           fmt.Sprintf("hn_%s", hit.ObjectID),
			Source:       "hackernews",
			Keyword:      keyword,
			Author:       hit.Author,
			DiscoveredAt: time.Now().UTC(),
		}

		// Parse created_at
		if t, err := time.Parse(time.RFC3339, hit.CreatedAt); err == nil {
			m.PublishedAt = t
		}

		if tag == "story" {
			m.Type = "post"
			m.Title = hit.Title
			m.Content = hit.StoryText
			m.URL = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		} else {
			m.Type = "comment"
			m.Title = fmt.Sprintf("Comment on: %s", hit.StoryTitle)
			m.Content = hit.CommentText
			m.URL = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		}

		mentions = append(mentions, m)
	}

	return mentions, nil
}
