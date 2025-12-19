package collector

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rebelice/mention-monitor/internal/models"
)

// PkgGoDev collects mentions from pkg.go.dev (importers)
type PkgGoDev struct{}

func (p *PkgGoDev) Name() string { return "pkggodev" }

func (p *PkgGoDev) Collect(ctx context.Context, keywords []string) ([]models.Mention, error) {
	var mentions []models.Mention

	for _, kw := range keywords {
		// Only process keywords that look like Go package paths
		if !strings.Contains(kw, "/") {
			continue
		}

		results, err := p.getImporters(ctx, kw)
		if err != nil {
			continue
		}
		mentions = append(mentions, results...)
	}

	return mentions, nil
}

func (p *PkgGoDev) getImporters(ctx context.Context, packagePath string) ([]models.Mention, error) {
	// Fetch the importers page
	pageURL := fmt.Sprintf("https://pkg.go.dev/%s?tab=importedby", url.QueryEscape(packagePath))

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
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
		return nil, fmt.Errorf("pkg.go.dev returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var mentions []models.Mention

	// Parse importers list
	doc.Find(".ImportedBy-list a").Each(func(i int, s *goquery.Selection) {
		importerPath := strings.TrimSpace(s.Text())
		href, exists := s.Attr("href")
		if !exists || importerPath == "" {
			return
		}

		importerURL := "https://pkg.go.dev" + href

		mentions = append(mentions, models.Mention{
			ID:           fmt.Sprintf("pkggodev_%s_%s", packagePath, importerPath),
			Source:       "pkggodev",
			Type:         "import",
			Keyword:      packagePath,
			Title:        fmt.Sprintf("Imported by %s", importerPath),
			Content:      fmt.Sprintf("Package %s imports %s", importerPath, packagePath),
			URL:          importerURL,
			Author:       importerPath,
			DiscoveredAt: time.Now().UTC(),
		})
	})

	return mentions, nil
}
