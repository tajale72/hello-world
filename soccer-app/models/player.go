package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Player struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Position  string             `bson:"position" json:"position"`
	AvgRating float64            `bson:"avgRating" json:"avgRating"`
}