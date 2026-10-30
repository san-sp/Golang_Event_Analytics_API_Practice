package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"pulsetrack/internal/tenant"
)

const apiKeyPrefix = "pt_live_"

// ErrInvalidAPIKey is returned when a key is unknown, expired,
// revoked, or belongs to a disabled tenant.
var ErrInvalidAPIKey = errors.New("invalid API key")

// RepositoryWriter defines the persistence operation needed by creation.
type RepositoryWriter interface {
	Create(ctx context.Context, doc Document) error
}

// RepositoryReader looks up API-key metadata by its hash.
type RepositoryReader interface {
	GetByHash(ctx context.Context, keyHash string) (Document, error)
}

// TenantReader retrieves the tenant associated with an API key.
type TenantReader interface {
	GetByID(ctx context.Context, id bson.ObjectID) (tenant.Document, error)
}

type RepositoryRevoker interface {
	Revoke(
		ctx context.Context,
		tenantID bson.ObjectID,
		id bson.ObjectID,
	) (bool, error)
}

// Service handles API-key generation, storage, and validation.
type Service struct {
	repository       RepositoryWriter
	keyReader        RepositoryReader
	tenantRepository TenantReader
	revoker          RepositoryRevoker
}

// NewService creates a service for generating API keys.
func NewService(repository RepositoryWriter) *Service {
	return &Service{repository: repository}
}

// NewValidationService creates a service capable of validating API keys.
func NewValidationService(
	keyReader RepositoryReader,
	tenantRepository TenantReader,
) *Service {
	return &Service{
		keyReader:        keyReader,
		tenantRepository: tenantRepository,
	}
}

// CreatedKey contains the raw key shown to the caller once,
// along with the metadata that was saved.
type CreatedKey struct {
	APIKey   string
	Metadata Document
}

// Create generates a random API key, stores only its hash,
// and returns the raw secret to the caller once.
func (s *Service) Create(
	ctx context.Context,
	tenantID bson.ObjectID,
	name string,
	expiresAt *time.Time,
) (CreatedKey, error) {
	if s.repository == nil {
		return CreatedKey{}, fmt.Errorf("API-key writer is not configured")
	}
	if tenantID.IsZero() {
		return CreatedKey{}, fmt.Errorf("tenant ID cannot be empty")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return CreatedKey{}, fmt.Errorf("API-key name cannot be empty")
	}

	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return CreatedKey{}, fmt.Errorf("API-key expiry must be in the future")
	}

	if expiresAt != nil {
		expiresAtUTC := expiresAt.UTC()
		expiresAt = &expiresAtUTC
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return CreatedKey{}, fmt.Errorf("generate API key: %w", err)
	}

	rawKey := apiKeyPrefix + base64.RawURLEncoding.EncodeToString(secret)

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	now := time.Now().UTC()
	doc := Document{
		ID:        bson.NewObjectID(),
		TenantID:  tenantID,
		KeyHash:   keyHash,
		Name:      name,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	if err := s.repository.Create(ctx, doc); err != nil {
		return CreatedKey{}, fmt.Errorf("save API-key metadata: %w", err)
	}

	return CreatedKey{
		APIKey:   rawKey,
		Metadata: doc,
	}, nil
}

// Validate checks the key hash, expiration, revocation, and tenant status.
// On success, it returns the trusted tenant ID.
func (s *Service) Validate(
	ctx context.Context,
	rawKey string,
) (bson.ObjectID, error) {
	if s.keyReader == nil || s.tenantRepository == nil {
		return bson.NilObjectID, fmt.Errorf("API-key validation is not configured")
	}

	rawKey = strings.TrimSpace(rawKey)
	if !strings.HasPrefix(rawKey, apiKeyPrefix) {
		return bson.NilObjectID, ErrInvalidAPIKey
	}

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	keyDoc, err := s.keyReader.GetByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return bson.NilObjectID, ErrInvalidAPIKey
		}
		return bson.NilObjectID, fmt.Errorf("look up API key: %w", err)
	}

	now := time.Now()

	if keyDoc.RevokedAt != nil {
		return bson.NilObjectID, ErrInvalidAPIKey
	}

	if keyDoc.ExpiresAt != nil && !keyDoc.ExpiresAt.After(now) {
		return bson.NilObjectID, ErrInvalidAPIKey
	}

	tenantDoc, err := s.tenantRepository.GetByID(ctx, keyDoc.TenantID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return bson.NilObjectID, ErrInvalidAPIKey
		}
		return bson.NilObjectID, fmt.Errorf("look up API-key tenant: %w", err)
	}

	if tenantDoc.Status != tenant.StatusActive {
		return bson.NilObjectID, ErrInvalidAPIKey
	}

	return keyDoc.TenantID, nil
}

// NewRevocationService creates a service for revoking API keys.
func NewRevocationService(revoker RepositoryRevoker) *Service {
	return &Service{
		revoker: revoker,
	}
}

// Revoke revokes an API key belonging to the specified tenant.
func (s *Service) Revoke(
	ctx context.Context,
	tenantID bson.ObjectID,
	keyID bson.ObjectID,
) (bool, error) {
	if s.revoker == nil {
		return false, fmt.Errorf("API-key revocation is not configured")
	}

	if tenantID.IsZero() {
		return false, fmt.Errorf("tenant ID cannot be empty")
	}

	if keyID.IsZero() {
		return false, fmt.Errorf("API-key ID cannot be empty")
	}

	revoked, err := s.revoker.Revoke(ctx, tenantID, keyID)
	if err != nil {
		return false, fmt.Errorf("revoke API key: %w", err)
	}

	return revoked, nil
}
