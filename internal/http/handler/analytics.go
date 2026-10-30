package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"pulsetrack/internal/analytics"
	"pulsetrack/internal/http/middleware"
)

type AnalyticsService interface {
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
	) ([]analytics.EventNameCount, error)

	CountEventsByDay(
		ctx context.Context,
		tenantID bson.ObjectID,
		from time.Time,
		to time.Time,
	) ([]analytics.DailyEventCount, error)
}

type AnalyticsHandler struct {
	service AnalyticsService
}

func NewAnalyticsHandler(service AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) CountEvents(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	fromValue := c.Query("from")
	toValue := c.Query("to")

	if fromValue == "" || toValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from and to query parameters are required",
		})
		return
	}

	from, err := time.Parse(time.RFC3339, fromValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be a valid RFC3339 timestamp",
		})
		return
	}

	to, err := time.Parse(time.RFC3339, toValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to must be a valid RFC3339 timestamp",
		})
		return
	}

	from, to, ok = parseTimeRange(c)
	if !ok {
		return
	}

	if !from.Before(to) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be earlier than to",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, err := h.service.CountEvents(ctx, tenantID, from, to)
	if err != nil {
		log.Printf("count analytics events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to count events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"count": count,
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
		},
	})
}

func (h *AnalyticsHandler) EventsBreakdown(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	fromValue := c.Query("from")
	toValue := c.Query("to")

	if fromValue == "" || toValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from and to query parameters are required",
		})
		return
	}

	from, err := time.Parse(time.RFC3339, fromValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be a valid RFC3339 timestamp",
		})
		return
	}

	to, err := time.Parse(time.RFC3339, toValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to must be a valid RFC3339 timestamp",
		})
		return
	}

	from, to, ok = parseTimeRange(c)
	if !ok {
		return
	}

	if !from.Before(to) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be earlier than to",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	results, err := h.service.CountEventsByName(
		ctx,
		tenantID,
		from,
		to,
	)
	if err != nil {
		log.Printf("get analytics event breakdown: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve event breakdown",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

func (h *AnalyticsHandler) EventsTimeSeries(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	from, to, ok := parseTimeRange(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	results, err := h.service.CountEventsByDay(ctx, tenantID, from, to)
	if err != nil {
		log.Printf("get analytics event time series: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve event time series",
		})
		return
	}

	data := make([]gin.H, 0, len(results))
	for _, result := range results {
		data = append(data, gin.H{
			"date":  result.Date.UTC().Format("2006-01-02"),
			"count": result.Count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func parseTimeRange(c *gin.Context) (time.Time, time.Time, bool) {
	fromValue := c.Query("from")
	toValue := c.Query("to")

	if fromValue == "" || toValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from and to query parameters are required",
		})
		return time.Time{}, time.Time{}, false
	}

	from, err := time.Parse(time.RFC3339, fromValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be a valid RFC3339 timestamp",
		})
		return time.Time{}, time.Time{}, false
	}

	to, err := time.Parse(time.RFC3339, toValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to must be a valid RFC3339 timestamp",
		})
		return time.Time{}, time.Time{}, false
	}

	from = from.UTC()
	to = to.UTC()

	if !from.Before(to) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be earlier than to",
		})
		return time.Time{}, time.Time{}, false
	}

	return from, to, true
}
