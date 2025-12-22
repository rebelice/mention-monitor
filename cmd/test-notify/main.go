package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rebelice/mention-monitor/internal/models"
	"github.com/rebelice/mention-monitor/internal/notifier"
)

func main() {
	ctx := context.Background()

	// Mock mention
	mention := models.Mention{
		ID:           "test_" + time.Now().Format("20060102150405"),
		Source:       "hackernews",
		Type:         "post",
		Keyword:      "lazypg",
		Title:        "Show HN: lazypg - A terminal UI for PostgreSQL",
		Content:      "I built a terminal UI for PostgreSQL inspired by lazygit. It supports vim keybindings, JSONB viewer, and more.",
		URL:          "https://news.ycombinator.com/item?id=12345678",
		Author:       "test_user",
		DiscoveredAt: time.Now().UTC(),
		PublishedAt:  time.Now().UTC(),
	}

	mentions := []models.Mention{mention}

	fmt.Println("Testing notifications with mock data...")
	fmt.Printf("Mock mention: %s - %s\n", mention.Source, mention.Title)

	// Test PostgreSQL (Supabase)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		fmt.Println("\nSending to PostgreSQL...")
		pg, err := notifier.NewPostgres(ctx, dbURL)
		if err != nil {
			fmt.Printf("PostgreSQL connection error: %v\n", err)
		} else {
			defer pg.Close()
			if err := pg.Send(ctx, mentions); err != nil {
				fmt.Printf("PostgreSQL error: %v\n", err)
			} else {
				fmt.Println("PostgreSQL: Success!")
			}
		}
	} else {
		fmt.Println("PostgreSQL: Skipped (not configured)")
	}

	// Test Bark
	barkKey := os.Getenv("BARK_DEVICE_KEY")
	barkServer := os.Getenv("BARK_SERVER_URL")
	if barkKey != "" {
		fmt.Println("\nSending Bark notification...")
		var bark *notifier.Bark
		if barkServer != "" {
			bark = notifier.NewBarkWithServer(barkServer, barkKey)
		} else {
			bark = notifier.NewBark(barkKey)
		}
		if err := bark.Send(ctx, mentions); err != nil {
			fmt.Printf("Bark error: %v\n", err)
		} else {
			fmt.Println("Bark: Success!")
		}
	} else {
		fmt.Println("Bark: Skipped (not configured)")
	}

	fmt.Println("\nTest complete!")
}
