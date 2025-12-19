package models

import "time"

// Mention represents a single mention of a keyword
type Mention struct {
	ID           string    `json:"id"`
	Source       string    `json:"source"`        // hackernews, reddit, github, twitter, devto, medium, stackoverflow, producthunt, lobsters, pkggodev, google
	Type         string    `json:"type"`          // post, comment, issue, discussion, article, question, answer
	Keyword      string    `json:"keyword"`       // matched keyword
	Title        string    `json:"title"`         // title or comment excerpt
	Content      string    `json:"content"`       // full content
	URL          string    `json:"url"`           // link to original
	Author       string    `json:"author"`        // author name
	DiscoveredAt time.Time `json:"discovered_at"` // when we found it
	PublishedAt  time.Time `json:"published_at"`  // when it was published (if available)
}

// Data represents the stored data structure
type Data struct {
	LastUpdated time.Time `json:"last_updated"`
	Mentions    []Mention `json:"mentions"`
}
