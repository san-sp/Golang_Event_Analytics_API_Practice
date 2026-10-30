package analytics

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EventNameCount struct {
	EventName string `json:"event_name" bson:"_id"`
	Count     int64  `json:"count" bson:"count"`
}

type DailyEventCount struct {
	Date  time.Time `json:"-" bson:"_id"`
	Count int64     `json:"count" bson:"count"`
}

// Repository performs analytics queries against stored events.
type Repository struct {
	events *mongo.Collection
}

// NewRepository creates an analytics repository.
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		events: db.Collection("events"),
	}
}

// CountEvents returns the number of events for a tenant
// in the half-open interval [from, to).
func (r *Repository) CountEvents(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) (int64, error) {
	if tenantID.IsZero() {
		return 0, fmt.Errorf("tenant ID cannot be empty")
	}

	if !from.Before(to) {
		return 0, fmt.Errorf("from must be earlier than to")
	}

	filter := bson.D{
		{Key: "tenant_id", Value: tenantID},
		{Key: "occurred_at", Value: bson.D{
			{Key: "$gte", Value: from.UTC()},
			{Key: "$lt", Value: to.UTC()},
		}},
	}

	count, err := r.events.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count tenant events: %w", err)
	}

	return count, nil
}

func (r *Repository) CountEventsByName(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) ([]EventNameCount, error) {
	if tenantID.IsZero() {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be earlier than to")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "tenant_id", Value: tenantID},
			{Key: "occurred_at", Value: bson.D{
				{Key: "$gte", Value: from.UTC()},
				{Key: "$lt", Value: to.UTC()},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$event_name"},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "count", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: 100}},
	}

	cursor, err := r.events.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate events by name: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]EventNameCount, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode event name counts: %w", err)
	}

	return results, nil
}

func (r *Repository) CountEventsByDay(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) ([]DailyEventCount, error) {
	if tenantID.IsZero() {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be earlier than to")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "tenant_id", Value: tenantID},
			{Key: "occurred_at", Value: bson.D{
				{Key: "$gte", Value: from.UTC()},
				{Key: "$lt", Value: to.UTC()},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$dateTrunc", Value: bson.D{
					{Key: "date", Value: "$occurred_at"},
					{Key: "unit", Value: "day"},
					{Key: "timezone", Value: "UTC"},
				}},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "_id", Value: 1},
		}}},
	}

	cursor, err := r.events.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate events by day: %w", err)
	}
	defer cursor.Close(ctx)

	results := make([]DailyEventCount, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decode daily event counts: %w", err)
	}

	return results, nil
}
