package tenant

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Status describes whether a tenant is allowed to use PulseTrack.
type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

// Document represents a customer workspace stored in MongoDB.
type Document struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	Status    Status        `bson:"status"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
