package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"pulsetrack/internal/event"
	"pulsetrack/internal/http/middleware"
)

type EventCreator interface {
	Create(ctx context.Context, doc event.Document) error
}

type EventLister interface {
	ListByTenant(
		ctx context.Context,
		tenantID bson.ObjectID,
		pageSize int64,
	) ([]event.Document, error)
}

// EventHandler handles event-ingestion requests.
type EventHandler struct {
	repository EventCreator
	creator    EventCreator
	lister     EventLister
}

// NewEventHandler creates an event handler.
func NewEventHandler(repository interface {
	EventCreator
	EventLister
}) *EventHandler {
	return &EventHandler{
		creator: repository,
		lister:  repository,
	}
}

type createEventRequest struct {
	EventName  string    `json:"event_name"`
	OccurredAt time.Time `json:"occurred_at"`
	ActorID    string    `json:"actor_id"`
	SessionID  string    `json:"session_id"`
	Properties bson.M    `json:"properties"`
}

// Create accepts and stores an event for the authenticated tenant.
func (h *EventHandler) Create(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid event request",
		})
		return
	}

	req.EventName = strings.TrimSpace(req.EventName)
	if req.EventName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "event_name is required",
		})
		return
	}

	if req.OccurredAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "occurred_at is required",
		})
		return
	}

	if len(req.EventName) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "event_name must not exceed 200 characters",
		})
		return
	}

	if len(req.ActorID) > 200 || len(req.SessionID) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "actor_id and session_id must not exceed 200 characters",
		})
		return
	}

	if req.Properties == nil {
		req.Properties = bson.M{}
	}

	doc := event.Document{
		ID:         bson.NewObjectID(),
		TenantID:   tenantID,
		EventName:  req.EventName,
		OccurredAt: req.OccurredAt.UTC(),
		ReceivedAt: time.Now().UTC(),
		ActorID:    strings.TrimSpace(req.ActorID),
		SessionID:  strings.TrimSpace(req.SessionID),
		Properties: req.Properties,
	}

	if err := h.creator.Create(c.Request.Context(), doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to store event",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
		"data": gin.H{
			"id":          doc.ID.Hex(),
			"event_name":  doc.EventName,
			"occurred_at": doc.OccurredAt,
			"received_at": doc.ReceivedAt,
		},
	})
}

// List returns the newest events for the authenticated tenant.
func (h *EventHandler) List(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	pageSize := int64(20)

	if rawLimit := c.Query("limit"); rawLimit != "" {
		limit, err := strconv.ParseInt(rawLimit, 10, 64)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		pageSize = limit
	}

	if pageSize > 100 {
		pageSize = 100
	}

	events, err := h.lister.ListByTenant(
		c.Request.Context(),
		tenantID,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve events",
		})
		return
	}

	items := make([]gin.H, 0, len(events))

	for _, doc := range events {
		items = append(items, gin.H{
			"id":          doc.ID.Hex(),
			"event_name":  doc.EventName,
			"occurred_at": doc.OccurredAt,
			"received_at": doc.ReceivedAt,
			"actor_id":    doc.ActorID,
			"session_id":  doc.SessionID,
			"properties":  doc.Properties,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"count": len(items),
	})
}
