package collector

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/rebelice/mention-monitor/internal/models"
)

// ProductHunt collects mentions from Product Hunt via RSS
type ProductHunt struct{}

func (p *ProductHunt) Name() string { return "producthunt" }

func (p *ProductHunt) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	// Product Hunt doesn't have keyword search RSS, so we fetch latest and filter
	feedURL := "https://www.producthunt.com/feed"

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
		return nil, fmt.Errorf("producthunt returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		// Check if any keyword matches
		text := item.Title + " " + item.Description
		if found, matchedKw := ContainsKeyword(text, keywords); found {
			mention := models.Mention{
				ID:           fmt.Sprintf("producthunt_%s", item.GUID),
				Source:       "producthunt",
				Type:         "post",
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
