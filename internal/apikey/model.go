package apikey

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Document stores API-key metadata. The raw API key must never be stored here.
type Document struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	TenantID  bson.ObjectID `bson:"tenant_id"`
	KeyHash   string        `bson:"key_hash"`
	Name      string        `bson:"name"`
	CreatedAt time.Time     `bson:"created_at"`
	ExpiresAt *time.Time    `bson:"expires_at,omitempty"`
	RevokedAt *time.Time    `bson:"revoked_at,omitempty"`
}
