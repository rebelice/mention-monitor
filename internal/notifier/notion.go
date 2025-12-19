package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rebelice/mention-monitor/internal/models"
)

// Notion sends mentions to a Notion database
type Notion struct {
	Token      string // Integration token
	DatabaseID string // Database ID to add pages to
}

// notionPage represents a page to create in Notion
type notionPage struct {
	Parent     notionParent     `json:"parent"`
	Properties notionProperties `json:"properties"`
}

type notionParent struct {
	DatabaseID string `json:"database_id"`
}

type notionProperties struct {
	Title      notionTitle  `json:"Title"`
	Source     notionSelect `json:"Source"`
	Type       notionSelect `json:"Type"`
	URL        notionURL    `json:"URL"`
	Author     notionText   `json:"Author"`
	Content    notionText   `json:"Content"`
	Keyword    notionSelect `json:"Keyword"`
	Discovered notionDate   `json:"Discovered"`
	Status     notionSelect `json:"Status"`
}

type notionTitle struct {
	Title []notionRichText `json:"title"`
}

type notionText struct {
	RichText []notionRichText `json:"rich_text"`
}

type notionRichText struct {
	Text notionTextContent `json:"text"`
}

type notionTextContent struct {
	Content string `json:"content"`
}

type notionSelect struct {
	Select notionSelectOption `json:"select"`
}

type notionSelectOption struct {
	Name string `json:"name"`
}

type notionURL struct {
	URL string `json:"url"`
}

type notionDate struct {
	Date notionDateValue `json:"date"`
}

type notionDateValue struct {
	Start string `json:"start"`
}

// NewNotion creates a new Notion notifier
func NewNotion(token, databaseID string) *Notion {
	return &Notion{
		Token:      token,
		DatabaseID: databaseID,
	}
}

// Send adds mentions to the Notion database
func (n *Notion) Send(ctx context.Context, mentions []models.Mention) error {
	if n.Token == "" || n.DatabaseID == "" {
		return fmt.Errorf("notion token or database ID not configured")
	}

	for _, m := range mentions {
		if err := n.createPage(ctx, m); err != nil {
			// Log error but continue with other mentions
			fmt.Printf("Failed to create Notion page for %s: %v\n", m.ID, err)
		}
	}

	return nil
}

func (n *Notion) createPage(ctx context.Context, m models.Mention) error {
	page := notionPage{
		Parent: notionParent{
			DatabaseID: n.DatabaseID,
		},
		Properties: notionProperties{
			Title: notionTitle{
				Title: []notionRichText{
					{Text: notionTextContent{Content: truncateString(m.Title, 2000)}},
				},
			},
			Source: notionSelect{
				Select: notionSelectOption{Name: m.Source},
			},
			Type: notionSelect{
				Select: notionSelectOption{Name: m.Type},
			},
			URL: notionURL{
				URL: m.URL,
			},
			Author: notionText{
				RichText: []notionRichText{
					{Text: notionTextContent{Content: m.Author}},
				},
			},
			Content: notionText{
				RichText: []notionRichText{
					{Text: notionTextContent{Content: truncateString(m.Content, 2000)}},
				},
			},
			Keyword: notionSelect{
				Select: notionSelectOption{Name: m.Keyword},
			},
			Discovered: notionDate{
				Date: notionDateValue{Start: m.DiscoveredAt.Format(time.RFC3339)},
			},
			Status: notionSelect{
				Select: notionSelectOption{Name: "unread"},
			},
		},
	}

	body, err := json.Marshal(page)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/pages", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+n.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("notion API error %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// CheckDuplicate checks if a mention already exists in the database
func (n *Notion) CheckDuplicate(ctx context.Context, url string) (bool, error) {
	if n.Token == "" || n.DatabaseID == "" {
		return false, nil
	}

	query := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "URL",
			"url": map[string]interface{}{
				"equals": url,
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", n.DatabaseID),
		bytes.NewReader(body))
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+n.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil // Assume not duplicate on error
	}

	var result struct {
		Results []interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return len(result.Results) > 0, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
