package notifier

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rebelice/mention-monitor/internal/models"
)

// Bark sends push notifications via Bark (iOS)
type Bark struct {
	// ServerURL is the Bark server URL (default: https://api.day.app)
	ServerURL string
	// DeviceKey is your Bark device key
	DeviceKey string
}

// NewBark creates a new Bark notifier
func NewBark(deviceKey string) *Bark {
	return &Bark{
		ServerURL: "https://api.day.app",
		DeviceKey: deviceKey,
	}
}

// NewBarkWithServer creates a new Bark notifier with custom server
func NewBarkWithServer(serverURL, deviceKey string) *Bark {
	return &Bark{
		ServerURL: strings.TrimSuffix(serverURL, "/"),
		DeviceKey: deviceKey,
	}
}

// Send sends a notification for each mention
func (b *Bark) Send(ctx context.Context, mentions []models.Mention) error {
	if b.DeviceKey == "" {
		return fmt.Errorf("bark device key not configured")
	}

	for _, m := range mentions {
		if err := b.sendOne(ctx, m); err != nil {
			// Log error but continue with other mentions
			fmt.Printf("Failed to send Bark notification for %s: %v\n", m.ID, err)
		}
	}

	return nil
}

func (b *Bark) sendOne(ctx context.Context, m models.Mention) error {
	// Format: 详细模式
	// 🔔 New mention on Hacker News
	// Title: Show HN: lazypg - Terminal UI for PostgreSQL
	// Author: someone

	title := fmt.Sprintf("New mention on %s", formatSourceName(m.Source))
	body := fmt.Sprintf("Title: %s", m.Title)
	if m.Author != "" {
		body += fmt.Sprintf("\nAuthor: %s", m.Author)
	}

	// Build URL with parameters
	// Format: https://api.day.app/{key}/{title}/{body}?url={url}&group={group}
	pushURL := fmt.Sprintf("%s/%s/%s/%s",
		b.ServerURL,
		b.DeviceKey,
		url.PathEscape(title),
		url.PathEscape(body),
	)

	// Add query parameters
	params := url.Values{}
	params.Set("url", m.URL)                    // Click to open original URL
	params.Set("group", "mention-monitor")      // Group notifications
	params.Set("icon", getSourceIcon(m.Source)) // Source icon

	pushURL += "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", pushURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bark returned status %d", resp.StatusCode)
	}

	return nil
}

// SendBatch sends a single aggregated notification for multiple mentions
func (b *Bark) SendBatch(ctx context.Context, mentions []models.Mention) error {
	if b.DeviceKey == "" {
		return fmt.Errorf("bark device key not configured")
	}

	if len(mentions) == 0 {
		return nil
	}

	if len(mentions) == 1 {
		return b.sendOne(ctx, mentions[0])
	}

	// Aggregated notification
	title := fmt.Sprintf("%d new mentions", len(mentions))

	var bodyParts []string
	for i, m := range mentions {
		if i >= 5 {
			bodyParts = append(bodyParts, fmt.Sprintf("... and %d more", len(mentions)-5))
			break
		}
		bodyParts = append(bodyParts, fmt.Sprintf("• [%s] %s", m.Source, truncateString(m.Title, 50)))
	}
	body := strings.Join(bodyParts, "\n")

	pushURL := fmt.Sprintf("%s/%s/%s/%s",
		b.ServerURL,
		b.DeviceKey,
		url.PathEscape(title),
		url.PathEscape(body),
	)

	params := url.Values{}
	params.Set("group", "mention-monitor")
	pushURL += "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", pushURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bark returned status %d", resp.StatusCode)
	}

	return nil
}

func formatSourceName(source string) string {
	names := map[string]string{
		"hackernews":    "Hacker News",
		"reddit":        "Reddit",
		"github":        "GitHub",
		"twitter":       "Twitter",
		"devto":         "Dev.to",
		"medium":        "Medium",
		"stackoverflow": "Stack Overflow",
		"producthunt":   "Product Hunt",
		"lobsters":      "Lobsters",
		"pkggodev":      "pkg.go.dev",
		"google":        "Google",
	}
	if name, ok := names[source]; ok {
		return name
	}
	return source
}

func getSourceIcon(source string) string {
	// Using emoji as icons (Bark supports custom icons via URL too)
	icons := map[string]string{
		"hackernews":    "https://news.ycombinator.com/favicon.ico",
		"reddit":        "https://www.reddit.com/favicon.ico",
		"github":        "https://github.com/favicon.ico",
		"twitter":       "https://twitter.com/favicon.ico",
		"devto":         "https://dev.to/favicon.ico",
		"medium":        "https://medium.com/favicon.ico",
		"stackoverflow": "https://stackoverflow.com/favicon.ico",
		"producthunt":   "https://www.producthunt.com/favicon.ico",
		"lobsters":      "https://lobste.rs/favicon.ico",
		"pkggodev":      "https://pkg.go.dev/favicon.ico",
		"google":        "https://www.google.com/favicon.ico",
	}
	if icon, ok := icons[source]; ok {
		return icon
	}
	return ""
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
