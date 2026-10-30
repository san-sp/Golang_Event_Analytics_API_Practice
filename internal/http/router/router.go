package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pulsetrack/internal/analytics"
	"pulsetrack/internal/apikey"
	"pulsetrack/internal/event"
	"pulsetrack/internal/http/handler"
	"pulsetrack/internal/http/middleware"
)

// New creates the Gin router and registers application routes.
func New(
	apiKeyValidationService *apikey.Service,
	eventRepository *event.Repository,
	analyticsService *analytics.Service,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.PrometheusMetrics())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "pulsetrack",
		})
	})

	eventHandler := handler.NewEventHandler(eventRepository)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	r.POST(
		"/v1/events",
		middleware.APIKeyAuth(apiKeyValidationService),
		eventHandler.Create,
	)

	r.GET(
		"/v1/events",
		middleware.APIKeyAuth(apiKeyValidationService),
		eventHandler.List,
	)

	r.GET(
		"/v1/analytics/events/count",
		middleware.APIKeyAuth(apiKeyValidationService),
		analyticsHandler.CountEvents,
	)

	r.GET(
		"/v1/analytics/events/breakdown",
		middleware.APIKeyAuth(apiKeyValidationService),
		analyticsHandler.EventsBreakdown,
	)

	r.GET(
		"/v1/analytics/events/timeseries",
		middleware.APIKeyAuth(apiKeyValidationService),
		analyticsHandler.EventsTimeSeries,
	)

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}
