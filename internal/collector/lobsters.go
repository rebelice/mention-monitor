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

// Lobsters collects mentions from Lobsters via API
type Lobsters struct{}

type lobstersStory struct {
	ShortID        string    `json:"short_id"`
	Title          string    `json:"title"`
	URL            string    `json:"url"`
	Description    string    `json:"description"`
	CommentsURL    string    `json:"comments_url"`
	SubmitterUser  string    `json:"submitter_user"`
	CreatedAt      time.Time `json:"created_at"`
	CommentCount   int       `json:"comment_count"`
}

func (l *Lobsters) Name() string { return "lobsters" }

func (l *Lobsters) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		results, err := l.search(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (l *Lobsters) search(ctx context.Context, keyword string) ([]models.Mention, error) {
	apiURL := fmt.Sprintf("https://lobste.rs/search.json?q=%s&what=stories&order=newest", url.QueryEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mention-monitor/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("lobsters returned status %d", resp.StatusCode)
	}

	var stories []lobstersStory
	if err := json.NewDecoder(resp.Body).Decode(&stories); err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, s := range stories {
		// Verify keyword match
		text := s.Title + " " + s.Description
		if found, matchedKw := ContainsKeyword(text, []string{keyword}); found {
			mentions = append(mentions, models.Mention{
				ID:           fmt.Sprintf("lobsters_%s", s.ShortID),
				Source:       "lobsters",
				Type:         "post",
				Keyword:      matchedKw,
				Title:        s.Title,
				Content:      s.Description,
				URL:          s.CommentsURL,
				Author:       s.SubmitterUser,
				DiscoveredAt: time.Now().UTC(),
				PublishedAt:  s.CreatedAt,
			})
		}
	}

	return mentions, nil
}
