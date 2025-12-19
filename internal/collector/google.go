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

// Google collects mentions via Google Alerts RSS
// User needs to create alerts at https://www.google.com/alerts and get the RSS URL
type Google struct {
	// AlertRSSURLs are the RSS feed URLs from Google Alerts
	// You can get these by creating alerts and selecting "Deliver to: RSS feed"
	AlertRSSURLs []string
}

func (g *Google) Name() string { return "google" }

func (g *Google) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	if len(g.AlertRSSURLs) == 0 {
		// If no alert URLs configured, try the basic search approach
		return g.searchBasic(ctx, keywords)
	}

	var mentions []models.Mention

	for _, feedURL := range g.AlertRSSURLs {
		results, err := g.fetchAlertFeed(ctx, feedURL, keywords)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (g *Google) fetchAlertFeed(ctx context.Context, feedURL string, keywords []string) ([]models.Mention, error) {
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
		return nil, fmt.Errorf("google alerts returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		// Determine which keyword matched
		text := item.Title + " " + item.Description
		_, matchedKw := ContainsKeyword(text, keywords)
		if matchedKw == "" && len(keywords) > 0 {
			matchedKw = keywords[0] // Default to first keyword for alert feeds
		}

		mention := models.Mention{
			ID:           fmt.Sprintf("google_%s", item.GUID),
			Source:       "google",
			Type:         "webpage",
			Keyword:      matchedKw,
			Title:        item.Title,
			Content:      truncate(item.Description, 500),
			URL:          item.Link,
			DiscoveredAt: time.Now().UTC(),
		}

		if item.PublishedParsed != nil {
			mention.PublishedAt = *item.PublishedParsed
		}

		mentions = append(mentions, mention)
	}

	return mentions, nil
}

// searchBasic uses Google News RSS as a fallback
func (g *Google) searchBasic(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		feedURL := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", url.QueryEscape(kw))

		req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "mention-monitor/1.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			continue
		}

		fp := gofeed.NewParser()
		feed, err := fp.Parse(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		for _, item := range feed.Items {
			mention := models.Mention{
				ID:           fmt.Sprintf("google_%s", item.GUID),
				Source:       "google",
				Type:         "news",
				Keyword:      kw,
				Title:        item.Title,
				Content:      truncate(item.Description, 500),
				URL:          item.Link,
				DiscoveredAt: time.Now().UTC(),
			}

			if item.PublishedParsed != nil {
				mention.PublishedAt = *item.PublishedParsed
			}

			mentions = append(mentions, mention)
		}
	}

	return mentions, nil
}
