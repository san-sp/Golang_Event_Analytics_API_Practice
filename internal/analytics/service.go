package analytics

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EventCounter defines the repository operation needed by the service.

type AnalyticsRepository interface {
	CountEvents(
		ctx context.Context,
		tenantID bson.ObjectID,
		from time.Time,
		to time.Time,
	) (int64, error)

	CountEventsByName(
		ctx context.Context,
		tenantID bson.ObjectID,
		from time.Time,
		to time.Time,
	) ([]EventNameCount, error)

	CountEventsByDay(
		ctx context.Context,
		tenantID bson.ObjectID,
		from time.Time,
		to time.Time,
	) ([]DailyEventCount, error)
}

// Service provides event analytics operations.
type Service struct {
	repository AnalyticsRepository
}

// NewService creates an analytics service.
func NewService(repository AnalyticsRepository) *Service {
	return &Service{repository: repository}
}

// CountEvents returns the number of events in the requested time range.
func (s *Service) CountEvents(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) (int64, error) {
	if s.repository == nil {
		return 0, fmt.Errorf("analytics repository is not configured")
	}

	if tenantID.IsZero() {
		return 0, fmt.Errorf("tenant ID cannot be empty")
	}

	if !from.Before(to) {
		return 0, fmt.Errorf("from must be earlier than to")
	}

	count, err := s.repository.CountEvents(ctx, tenantID, from, to)
	if err != nil {
		return 0, fmt.Errorf("count events: %w", err)
	}

	return count, nil
}

func (s *Service) CountEventsByName(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) ([]EventNameCount, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("analytics repository is not configured")
	}
	if tenantID.IsZero() {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be earlier than to")
	}

	results, err := s.repository.CountEventsByName(
		ctx,
		tenantID,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("count events by name: %w", err)
	}

	return results, nil
}

func (s *Service) CountEventsByDay(
	ctx context.Context,
	tenantID bson.ObjectID,
	from time.Time,
	to time.Time,
) ([]DailyEventCount, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("analytics repository is not configured")
	}
	if tenantID.IsZero() {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be earlier than to")
	}

	results, err := s.repository.CountEventsByDay(
		ctx,
		tenantID,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf("count events by day: %w", err)
	}

	return results, nil
}
