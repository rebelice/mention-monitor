package notifier

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/rebelice/mention-monitor/internal/models"
)

// MongoDB stores mentions in MongoDB
type MongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// MentionDocument represents a mention document in MongoDB
type MentionDocument struct {
	ID           string    `bson:"id"`
	Source       string    `bson:"source"`
	Type         string    `bson:"type"`
	Keyword      string    `bson:"keyword"`
	Title        string    `bson:"title"`
	Content      string    `bson:"content"`
	URL          string    `bson:"url"`
	Author       string    `bson:"author"`
	DiscoveredAt time.Time `bson:"discovered_at"`
	PublishedAt  time.Time `bson:"published_at"`
	Status       string    `bson:"status"`
	CreatedAt    time.Time `bson:"created_at"`
}

// NewMongoDB creates a new MongoDB notifier
func NewMongoDB(uri string) (*MongoDB, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	collection := client.Database("mention_monitor").Collection("mentions")

	// Create indexes
	if err := createIndexes(ctx, collection); err != nil {
		fmt.Printf("Warning: failed to create indexes: %v\n", err)
	}

	return &MongoDB{
		client:     client,
		collection: collection,
	}, nil
}

func createIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "url", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "discovered_at", Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Send stores mentions in MongoDB
func (m *MongoDB) Send(ctx context.Context, mentions []models.Mention) error {
	if len(mentions) == 0 {
		return nil
	}

	for _, mention := range mentions {
		doc := MentionDocument{
			ID:           mention.ID,
			Source:       mention.Source,
			Type:         mention.Type,
			Keyword:      mention.Keyword,
			Title:        mention.Title,
			Content:      mention.Content,
			URL:          mention.URL,
			Author:       mention.Author,
			DiscoveredAt: mention.DiscoveredAt,
			PublishedAt:  mention.PublishedAt,
			Status:       "unread",
			CreatedAt:    time.Now().UTC(),
		}

		_, err := m.collection.InsertOne(ctx, doc)
		if err != nil {
			// Check if it's a duplicate key error, if so, skip
			if mongo.IsDuplicateKeyError(err) {
				fmt.Printf("Skipping duplicate mention: %s\n", mention.ID)
				continue
			}
			fmt.Printf("Failed to insert mention %s: %v\n", mention.ID, err)
		}
	}

	return nil
}

// CheckDuplicate checks if a mention with the given ID already exists
func (m *MongoDB) CheckDuplicate(ctx context.Context, id string) (bool, error) {
	count, err := m.collection.CountDocuments(ctx, bson.D{{Key: "id", Value: id}})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
