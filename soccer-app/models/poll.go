package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollDate  string             `bson:"pollDate" json:"pollDate"` // e.g. "2025-12-13"
	Status    string             `bson:"status" json:"status"`     // OPEN | CLOSED
	EndsAt    time.Time          `bson:"endsAt" json:"endsAt"`     // deadline
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type PollTeams struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	PollID      primitive.ObjectID `bson:"pollId"`
	TeamA       []User             `bson:"teamA"`
	TeamB       []User             `bson:"teamB"`
	YesCount    int                `bson:"yesCount"`
	GeneratedAt time.Time          `bson:"generatedAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
}
