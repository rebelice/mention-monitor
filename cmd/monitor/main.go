package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rebelice/mention-monitor/internal/collector"
	"github.com/rebelice/mention-monitor/internal/models"
	"github.com/rebelice/mention-monitor/internal/notifier"
)

const dataFile = "data/mentions.json"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Load configuration from environment
	config := loadConfig()

	fmt.Println("Starting mention monitor...")
	fmt.Printf("Keywords: %v\n", config.Keywords)

	// Initialize collectors
	coll := collector.New(
		&collector.HackerNews{},
		&collector.Reddit{},
		&collector.GitHub{Token: config.GitHubToken},
		&collector.Twitter{NitterInstances: collector.DefaultNitterInstances},
		&collector.DevTo{},
		&collector.Medium{},
		&collector.StackOverflow{},
		&collector.ProductHunt{},
		&collector.Lobsters{},
		&collector.PkgGoDev{},
		&collector.Google{AlertRSSURLs: config.GoogleAlertURLs},
	)

	// Load existing data
	data := loadData()
	seen := make(map[string]bool)
	for _, m := range data.Mentions {
		seen[m.ID] = true
	}

	fmt.Printf("Loaded %d existing mentions\n", len(data.Mentions))

	// Collect new mentions
	fmt.Println("Collecting mentions from all sources...")
	allMentions := coll.CollectAll(ctx, config.Keywords)
	fmt.Printf("Collected %d total mentions\n", len(allMentions))

	// Filter new mentions
	var newMentions []models.Mention
	for _, m := range allMentions {
		if !seen[m.ID] {
			seen[m.ID] = true
			newMentions = append(newMentions, m)
			data.Mentions = append(data.Mentions, m)
		}
	}

	fmt.Printf("Found %d new mentions\n", len(newMentions))

	if len(newMentions) > 0 {
		// Send to PostgreSQL (Supabase)
		if config.DatabaseURL != "" {
			fmt.Println("Sending to PostgreSQL...")
			pg, err := notifier.NewPostgres(ctx, config.DatabaseURL)
			if err != nil {
				fmt.Printf("PostgreSQL connection error: %v\n", err)
			} else {
				defer pg.Close()
				if err := pg.Send(ctx, newMentions); err != nil {
					fmt.Printf("PostgreSQL error: %v\n", err)
				} else {
					fmt.Printf("Added %d mentions to PostgreSQL\n", len(newMentions))
				}
			}
		}

		// Send Bark notification
		if config.BarkDeviceKey != "" {
			fmt.Println("Sending Bark notifications...")
			var bark *notifier.Bark
			if config.BarkServerURL != "" {
				bark = notifier.NewBarkWithServer(config.BarkServerURL, config.BarkDeviceKey)
			} else {
				bark = notifier.NewBark(config.BarkDeviceKey)
			}

			// Send individual notifications for detailed mode
			if err := bark.Send(ctx, newMentions); err != nil {
				fmt.Printf("Bark error: %v\n", err)
			} else {
				fmt.Printf("Sent %d Bark notifications\n", len(newMentions))
			}
		}
	}

	// Save data
	data.LastUpdated = time.Now().UTC()
	saveData(data)
	fmt.Println("Data saved successfully")
}

type Config struct {
	Keywords        []string
	GitHubToken     string
	GoogleAlertURLs []string
	DatabaseURL     string
	BarkDeviceKey   string
	BarkServerURL   string
}

func loadConfig() Config {
	keywords := os.Getenv("KEYWORDS")
	if keywords == "" {
		keywords = "lazypg,rebelice/lazypg"
	}

	googleAlerts := os.Getenv("GOOGLE_ALERT_URLS")
	var alertURLs []string
	if googleAlerts != "" {
		alertURLs = strings.Split(googleAlerts, ",")
	}

	return Config{
		Keywords:        strings.Split(keywords, ","),
		GitHubToken:     os.Getenv("GITHUB_TOKEN"),
		GoogleAlertURLs: alertURLs,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		BarkDeviceKey:   os.Getenv("BARK_DEVICE_KEY"),
		BarkServerURL:   os.Getenv("BARK_SERVER_URL"),
	}
}

func loadData() models.Data {
	data := models.Data{Mentions: []models.Mention{}}

	f, err := os.Open(dataFile)
	if err != nil {
		return data
	}
	defer f.Close()

	json.NewDecoder(f).Decode(&data)
	return data
}

func saveData(data models.Data) {
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return
	}

	f, err := os.Create(dataFile)
	if err != nil {
		fmt.Printf("Error saving data: %v\n", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Printf("Error encoding data: %v\n", err)
	}
}
