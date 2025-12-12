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
	Rating    int                `bson:"rating" json:"rating"`
	Attending bool               `bson:"attending" json:"attending"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
