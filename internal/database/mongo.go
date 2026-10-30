package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect creates a reusable MongoDB client and verifies that the
// database server is reachable before the API starts accepting requests.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	// Configure the driver with the connection string.
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("create MongoDB client: %w", err)
	}

	// Bound the initial ping so a database outage cannot stall startup
	// indefinitely. The caller controls the overall connection deadline.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		// Release client resources if the database isn't reachable.
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	return client, nil
}
