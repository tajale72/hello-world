package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    primitive.ObjectID `bson:"pollId" json:"pollId"`
	FirstName string             `bson:"firstName" json:"firstName"`
	LastName  string             `bson:"lastName" json:"lastName"`
	Secret    string             `bson:"secret" json:"secret"`
	Rating    int                `bson:"rating" json:"rating"`
	Attending bool               `bson:"attending" json:"attending"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
