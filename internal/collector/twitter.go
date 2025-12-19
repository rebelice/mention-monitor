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

// Twitter collects mentions from Twitter via Nitter RSS (free alternative)
type Twitter struct {
	// NitterInstance is the Nitter instance to use (e.g., "nitter.net")
	// Multiple instances can be tried as fallback
	NitterInstances []string
}

// DefaultNitterInstances are public Nitter instances
var DefaultNitterInstances = []string{
	"nitter.privacydev.net",
	"nitter.poast.org",
	"nitter.woodland.cafe",
}

func (t *Twitter) Name() string { return "twitter" }

func (t *Twitter) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	instances := t.NitterInstances
	if len(instances) == 0 {
		instances = DefaultNitterInstances
	}

	var mentions []models.Mention

	for _, kw := range keywords {
		for _, instance := range instances {
			results, err := t.search(ctx, instance, kw)
			if err != nil {
				// Try next instance
				continue
			}
			mentions = append(mentions, results...)
			break // Success, no need to try other instances
		}
	}

	return mentions, nil
}

func (t *Twitter) search(ctx context.Context, instance, keyword string) ([]models.Mention, error) {
	feedURL := fmt.Sprintf("https://%s/search/rss?f=tweets&q=%s", instance, url.QueryEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mention-monitor/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("nitter returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		m := models.Mention{
			ID:           fmt.Sprintf("twitter_%s", item.GUID),
			Source:       "twitter",
			Type:         "post",
			Keyword:      keyword,
			Title:        truncate(item.Title, 100),
			Content:      item.Description,
			URL:          convertNitterToTwitterURL(item.Link),
			DiscoveredAt: time.Now().UTC(),
		}

		if item.Author != nil {
			m.Author = item.Author.Name
		}
		if item.PublishedParsed != nil {
			m.PublishedAt = *item.PublishedParsed
		}

		mentions = append(mentions, m)
	}

	return mentions, nil
}

// convertNitterToTwitterURL converts Nitter URL back to Twitter URL
func convertNitterToTwitterURL(nitterURL string) string {
	u, err := url.Parse(nitterURL)
	if err != nil {
		return nitterURL
	}
	u.Host = "twitter.com"
	return u.String()
}
