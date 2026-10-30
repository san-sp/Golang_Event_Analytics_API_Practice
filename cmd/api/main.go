package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsetrack/internal/analytics"
	"pulsetrack/internal/apikey"
	"pulsetrack/internal/config"
	"pulsetrack/internal/database"
	"pulsetrack/internal/event"
	"pulsetrack/internal/http/middleware"
	"pulsetrack/internal/http/router"
	"pulsetrack/internal/tenant"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Give MongoDB a bounded startup window.
	connectCtx, cancelConnect := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	client, err := database.Connect(connectCtx, cfg.MongoURI)
	cancelConnect()

	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	// Create required indexes before serving HTTP requests.
	indexCtx, cancelIndexes := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)

	err = database.EnsureIndexes(
		indexCtx,
		client.Database(cfg.MongoDatabase),
	)

	cancelIndexes()

	if err != nil {
		log.Fatalf("failed to initialize MongoDB indexes: %v", err)
	}

	log.Println("MongoDB indexes initialized")

	// Ensure the database client is released when main returns.
	defer func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect MongoDB: %v", err)
		}
	}()

	// Initialize repositories.
	db := client.Database(cfg.MongoDatabase)

	tenantRepository := tenant.NewRepository(db)
	apiKeyRepository := apikey.NewRepository(db)
	eventRepository := event.NewRepository(db)

	// Initialize API-key services.
	apiKeyCreationService := apikey.NewService(apiKeyRepository)

	apiKeyValidationService := apikey.NewValidationService(
		apiKeyRepository,
		tenantRepository,
	)

	apiKeyRevocationService := apikey.NewRevocationService(apiKeyRepository)

	// Keep the event repository available for the event service
	// and HTTP handlers we will implement next.
	_ = eventRepository

	// Keep service references ready for handler wiring.
	_ = apiKeyCreationService
	_ = apiKeyValidationService
	_ = apiKeyRevocationService

	analyticsRepository := analytics.NewRepository(db)
	analyticsService := analytics.NewService(analyticsRepository)

	apiRouter := router.New(
		apiKeyValidationService,
		eventRepository,
		analyticsService,
	)

	// Step 7: Check whether MongoDB is currently reachable.
	// This endpoint is separate from /healthz because a running HTTP
	// process does not guarantee that database-dependent requests work.
	apiRouter.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			c.Request.Context(),
			2*time.Second,
		)
		defer cancel()

		if err := client.Ping(ctx, nil); err != nil {
			middleware.SetMongoDBReady(false)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"reason": "database_unavailable",
			})
			return
		}

		middleware.SetMongoDBReady(true)

		c.JSON(http.StatusOK, gin.H{
			"status":   "ready",
			"database": "connected",
		})
	})

	// Gracefully shut down when the process receives an interrupt
	// or termination signal.
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apiRouter,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		log.Printf("received signal %s; shutting down", sig)
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server failed: %v", err)
		}
		return
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}
}
