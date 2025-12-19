package collector

import (
	"context"
	"strings"

	"github.com/rebelice/mention-monitor/internal/models"
)

// Source defines the interface for all data sources
type Source interface {
	Name() string
	Collect(ctx context.Context, keywords []string) ([]models.Mention, error)
}

// Collector aggregates multiple sources
type Collector struct {
	sources []Source
}

// New creates a new collector with the given sources
func New(sources ...Source) *Collector {
	return &Collector{sources: sources}
}

// CollectAll collects mentions from all sources
func (c *Collector) CollectAll(ctx context.Context, keywords []string) []models.Mention {
	var all []models.Mention
	for _, src := range c.sources {
		mentions, err := src.Collect(ctx, keywords)
		if err != nil {
			// Log error but continue with other sources
			continue
		}
		all = append(all, mentions...)
	}
	return all
}

// ContainsKeyword checks if text contains any of the keywords (case-insensitive)
func ContainsKeyword(text string, keywords []string) (bool, string) {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true, kw
		}
	}
	return false, ""
}
