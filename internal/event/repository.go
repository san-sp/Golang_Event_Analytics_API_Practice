package event

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// Repository handles persistence for activity events.
type Repository struct {
	collection *mongo.Collection
}

// NewRepository creates an event repository using the provided database.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("events"),
	}
}

// Create stores an event belonging to a tenant.
func (r *Repository) Create(ctx context.Context, doc Document) error {
	if doc.TenantID.IsZero() {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	doc.EventName = strings.TrimSpace(doc.EventName)
	if doc.EventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}

	if doc.ReceivedAt.IsZero() {
		return fmt.Errorf("received timestamp cannot be empty")
	}

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

// ListByTenant returns events for one tenant, newest first.
// The tenant ID is always included in the database filter.
func (r *Repository) ListByTenant(
	ctx context.Context,
	tenantID bson.ObjectID,
	pageSize int64,
) ([]Document, error) {
	if tenantID.IsZero() {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	filter := bson.D{
		{Key: "tenant_id", Value: tenantID},
	}

	findOptions := options.Find().
		SetSort(bson.D{
			{Key: "occurred_at", Value: -1},
			{Key: "_id", Value: -1},
		}).
		SetLimit(pageSize)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find tenant events: %w", err)
	}
	defer cursor.Close(ctx)

	events := make([]Document, 0, pageSize)
	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf("decode tenant events: %w", err)
	}

	return events, nil
}
