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

	// Test MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		fmt.Println("\nSending to MongoDB...")
		mongodb, err := notifier.NewMongoDB(mongoURI)
		if err != nil {
			fmt.Printf("MongoDB connection error: %v\n", err)
		} else {
			defer mongodb.Close(ctx)
			if err := mongodb.Send(ctx, mentions); err != nil {
				fmt.Printf("MongoDB error: %v\n", err)
			} else {
				fmt.Println("MongoDB: Success!")
			}
		}
	} else {
		fmt.Println("MongoDB: Skipped (not configured)")
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
