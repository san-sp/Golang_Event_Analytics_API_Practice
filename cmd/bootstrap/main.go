package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"pulsetrack/internal/apikey"
	"pulsetrack/internal/config"
	"pulsetrack/internal/database"
	"pulsetrack/internal/tenant"
)

func main() {
	tenantName := strings.TrimSpace(os.Getenv("BOOTSTRAP_TENANT_NAME"))
	if tenantName == "" {
		log.Fatal("BOOTSTRAP_TENANT_NAME must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	client, err := database.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("disconnect MongoDB: %v", err)
		}
	}()

	db := client.Database(cfg.MongoDatabase)

	if err := database.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("initialize MongoDB indexes: %v", err)
	}

	tenantRepository := tenant.NewRepository(db)
	apiKeyRepository := apikey.NewRepository(db)
	apiKeyService := apikey.NewService(apiKeyRepository)

	tenantDoc, err := tenantRepository.Create(ctx, tenantName)
	if err != nil {
		log.Fatalf("create tenant: %v", err)
	}

	createdKey, err := apiKeyService.Create(
		ctx,
		tenantDoc.ID,
		"initial-bootstrap-key",
		nil,
	)
	if err != nil {
		log.Fatalf("create initial API key: %v", err)
	}

	fmt.Println("Tenant created successfully.")
	fmt.Printf("Tenant ID: %s\n", tenantDoc.ID.Hex())
	fmt.Println("API key created successfully.")
	fmt.Printf("API key: %s\n", createdKey.APIKey)
	fmt.Println("Store this key securely; it cannot be retrieved in plaintext later.")
}
