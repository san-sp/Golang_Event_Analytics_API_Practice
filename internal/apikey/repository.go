package apikey

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Repository handles persistence for API-key metadata.
type Repository struct {
	collection *mongo.Collection
}

// NewRepository creates an API-key repository using the provided database.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("api_keys"),
	}
}

// Create stores API-key metadata. The raw API key must never be passed here.
func (r *Repository) Create(ctx context.Context, doc Document) error {
	if doc.TenantID.IsZero() {
		return fmt.Errorf("tenant ID cannot be empty")
	}

	doc.KeyHash = strings.TrimSpace(doc.KeyHash)
	if doc.KeyHash == "" {
		return fmt.Errorf("API-key hash cannot be empty")
	}

	doc.Name = strings.TrimSpace(doc.Name)
	if doc.Name == "" {
		return fmt.Errorf("API-key name cannot be empty")
	}

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert API key: %w", err)
	}

	return nil
}

// GetByHash retrieves API-key metadata using its cryptographic hash.
func (r *Repository) GetByHash(
	ctx context.Context,
	keyHash string,
) (Document, error) {
	keyHash = strings.TrimSpace(keyHash)
	if keyHash == "" {
		return Document{}, fmt.Errorf("API-key hash cannot be empty")
	}

	var doc Document

	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "key_hash", Value: keyHash}},
	).Decode(&doc)
	if err != nil {
		return Document{}, fmt.Errorf("find API key by hash: %w", err)
	}

	return doc, nil
}

// Revoke marks an API key as revoked without deleting its record.
// It only revokes a key belonging to the specified tenant.
// The bool is false if the key doesn't exist, belongs to another tenant,
// or has already been revoked.
func (r *Repository) Revoke(
	ctx context.Context,
	tenantID bson.ObjectID,
	id bson.ObjectID,
) (bool, error) {
	if tenantID.IsZero() {
		return false, fmt.Errorf("tenant ID cannot be empty")
	}

	if id.IsZero() {
		return false, fmt.Errorf("API-key ID cannot be empty")
	}

	now := time.Now().UTC()

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "tenant_id", Value: tenantID},
		{Key: "revoked_at", Value: bson.D{
			{Key: "$exists", Value: false},
		}},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "revoked_at", Value: now},
		}},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("revoke API key: %w", err)
	}

	return result.MatchedCount == 1, nil
}
