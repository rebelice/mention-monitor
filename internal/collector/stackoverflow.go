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

// StackOverflow collects mentions from Stack Overflow via RSS
type StackOverflow struct{}

func (s *StackOverflow) Name() string { return "stackoverflow" }

func (s *StackOverflow) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		results, err := s.search(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (s *StackOverflow) search(ctx context.Context, keyword string) ([]models.Mention, error) {
	// Stack Overflow search RSS
	feedURL := fmt.Sprintf("https://stackoverflow.com/feeds/tag/%s", url.QueryEscape(keyword))

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
		// Try alternative search URL
		return s.searchAlternative(ctx, keyword)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		mention := models.Mention{
			ID:           fmt.Sprintf("stackoverflow_%s", item.GUID),
			Source:       "stackoverflow",
			Type:         "question",
			Keyword:      keyword,
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

	return mentions, nil
}

func (s *StackOverflow) searchAlternative(ctx context.Context, keyword string) ([]models.Mention, error) {
	// Alternative: search RSS
	feedURL := fmt.Sprintf("https://stackoverflow.com/feeds/search?q=%s", url.QueryEscape(keyword))

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
		return nil, fmt.Errorf("stackoverflow returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		// Verify keyword is in content
		text := item.Title + " " + item.Description
		if found, matchedKw := ContainsKeyword(text, []string{keyword}); !found {
			continue
		} else {
			mention := models.Mention{
				ID:           fmt.Sprintf("stackoverflow_%s", item.GUID),
				Source:       "stackoverflow",
				Type:         "question",
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
