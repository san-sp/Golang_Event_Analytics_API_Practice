package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes creates the indexes required by PulseTrack.
// It should run during startup before the HTTP server begins accepting requests.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	indexCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	indexes := map[string][]mongo.IndexModel{
		"api_keys": {
			{
				Keys: bson.D{
					{Key: "key_hash", Value: 1},
				},
				Options: options.Index().
					SetUnique(true).
					SetName("uq_api_keys_key_hash"),
			},
		},
		"events": {
			{
				Keys: bson.D{
					{Key: "tenant_id", Value: 1},
					{Key: "occurred_at", Value: -1},
					{Key: "_id", Value: -1},
				},
				Options: options.Index().
					SetName("idx_events_tenant_occurred_id"),
			},
			{
				Keys: bson.D{
					{Key: "tenant_id", Value: 1},
					{Key: "event_name", Value: 1},
					{Key: "occurred_at", Value: -1},
					{Key: "_id", Value: -1},
				},
				Options: options.Index().
					SetName("idx_events_tenant_name_occurred_id"),
			},
		},
	}

	for collectionName, models := range indexes {
		collection := db.Collection(collectionName)

		if _, err := collection.Indexes().CreateMany(indexCtx, models); err != nil {
			return fmt.Errorf("create indexes for %s: %w", collectionName, err)
		}
	}

	return nil
}
