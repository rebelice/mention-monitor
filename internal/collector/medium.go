package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rebelice/mention-monitor/internal/models"
)

// Medium collects mentions from Medium via RSS
type Medium struct{}

func (m *Medium) Name() string { return "medium" }

func (m *Medium) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		results, err := m.search(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (m *Medium) search(ctx context.Context, keyword string) ([]models.Mention, error) {
	// Medium's tag-based RSS feed
	feedURL := fmt.Sprintf("https://medium.com/feed/tag/%s", url.QueryEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
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
		return nil, fmt.Errorf("medium returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		// Additional keyword check in title/content
		text := item.Title + " " + item.Description
		if found, matchedKw := ContainsKeyword(text, []string{keyword}); !found {
			continue
		} else {
			mention := models.Mention{
				ID:           fmt.Sprintf("medium_%s", item.GUID),
				Source:       "medium",
				Type:         "article",
				Keyword:      matchedKw,
				Title:        item.Title,
				Content:      truncate(item.Description, 500),
				URL:          item.Link,
				DiscoveredAt: time.Now().UTC(),
			}

			if item.Author != nil {
				mention.Author = item.Author.Name
			}
			if item.PublishedParsed != nil {
				mention.PublishedAt = *item.PublishedParsed
			}

			mentions = append(mentions, mention)
		}
	}

	return mentions, nil
}
