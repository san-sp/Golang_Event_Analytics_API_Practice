package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsetrack_http_requests_total",
			Help: "Total number of HTTP requests handled by PulseTrack.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsetrack_http_request_duration_seconds",
			Help:    "Duration of HTTP requests handled by PulseTrack.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	mongodbReady = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pulsetrack_mongodb_ready",
			Help: "Whether MongoDB is reachable (1 = ready, 0 = unavailable).",
		},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		mongodbReady,
	)
}

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(
			method,
			route,
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			method,
			route,
			status,
		).Observe(duration)
	}
}

func SetMongoDBReady(ready bool) {
	if ready {
		mongodbReady.Set(1)
		return
	}

	mongodbReady.Set(0)
}
