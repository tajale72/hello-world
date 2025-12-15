package models

import "time"

type User struct {
	FirstName  string    `bson:"firstName" json:"firstName"`
	LastName   string    `bson:"lastName" json:"lastName"`
	Position   string    `bson:"position" json:"position"`
	Skills     []Skill   `bson:"skills" json:"skills"`
	SecretHash string    `bson:"secretHash"`
	CreatedAt  time.Time `bson:"createdAt"`
}

type Skill struct {
	Name  string `json:"name" bson:"name"`
	Value int    `json:"value" bson:"value"`
}
