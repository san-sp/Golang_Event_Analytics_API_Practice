package event

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Document represents one activity event stored by PulseTrack.
//
// TenantID and ReceivedAt are server-owned fields. They must be assigned
// by trusted application code, never accepted from an untrusted client.
type Document struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	TenantID   bson.ObjectID `bson:"tenant_id"`
	EventName  string        `bson:"event_name"`
	OccurredAt time.Time     `bson:"occurred_at"`
	ReceivedAt time.Time     `bson:"received_at"`
	ActorID    string        `bson:"actor_id,omitempty"`
	SessionID  string        `bson:"session_id,omitempty"`
	Properties bson.M        `bson:"properties,omitempty"`
}
