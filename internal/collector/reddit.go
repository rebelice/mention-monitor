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

// Reddit collects mentions from Reddit via RSS
type Reddit struct{}

func (r *Reddit) Name() string { return "reddit" }

func (r *Reddit) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		// Search posts
		posts, err := r.searchPosts(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, posts...)

		// Search comments
		comments, err := r.searchComments(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, comments...)
	}

	return mentions, nil
}

func (r *Reddit) searchPosts(ctx context.Context, keyword string) ([]models.Mention, error) {
	feedURL := fmt.Sprintf("https://www.reddit.com/search.rss?q=%s&sort=new&t=day", url.QueryEscape(keyword))
	return r.fetch(ctx, feedURL, keyword, "post")
}

func (r *Reddit) searchComments(ctx context.Context, keyword string) ([]models.Mention, error) {
	feedURL := fmt.Sprintf("https://www.reddit.com/search.rss?q=%s&sort=new&t=day&type=comment", url.QueryEscape(keyword))
	return r.fetch(ctx, feedURL, keyword, "comment")
}

func (r *Reddit) fetch(ctx context.Context, feedURL, keyword, contentType string) ([]models.Mention, error) {
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
		return nil, fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention
	for _, item := range feed.Items {
		m := models.Mention{
			ID:           fmt.Sprintf("reddit_%s", item.GUID),
			Source:       "reddit",
			Type:         contentType,
			Keyword:      keyword,
			Title:        item.Title,
			Content:      item.Description,
			URL:          item.Link,
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
