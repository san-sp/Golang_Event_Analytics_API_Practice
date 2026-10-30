package tenant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Repository handles persistence for tenant documents.
type Repository struct {
	collection *mongo.Collection
}

// NewRepository creates a tenant repository using the provided database.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("tenants"),
	}
}

// Create stores a new tenant with a server-generated ID and timestamps.
func (r *Repository) Create(ctx context.Context, name string) (Document, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Document{}, fmt.Errorf("tenant name cannot be empty")
	}

	now := time.Now().UTC()

	doc := Document{
		ID:        bson.NewObjectID(),
		Name:      name,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return Document{}, fmt.Errorf("insert tenant: %w", err)
	}

	return doc, nil
}

// GetByID retrieves a tenant by its MongoDB ObjectID.
func (r *Repository) GetByID(
	ctx context.Context,
	id bson.ObjectID,
) (Document, error) {
	var doc Document

	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
	).Decode(&doc)
	if err != nil {
		return Document{}, fmt.Errorf("find tenant: %w", err)
	}

	return doc, nil
}
