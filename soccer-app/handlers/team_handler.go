package handlers

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"soccer-app/models"
)

type PlayerRating struct {
	Name     string         `json:"name"`
	Skills   []models.Skill `json:"skills,omitempty"`
	Position string         `json:"position,omitempty"`
	Rating   float64        `json:"rating"`
}

type voteDoc struct {
	FirstName string             `bson:"firstName"`
	LastName  string             `bson:"lastName"`
	Rating    int                `bson:"rating"`
	Attending bool               `bson:"attending"`
	PollID    primitive.ObjectID `bson:"pollId"`
}

type TeamPlayer struct {
	UserID    primitive.ObjectID `json:"userId"`
	FirstName string             `json:"firstName"`
	LastName  string             `json:"lastName"`
	Position  string             `json:"position"`
	Skills    []models.Skill     `json:"skills"`
}

func GenerateTeams(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		pollHex := c.Param("id")
		pollOID, err := primitive.ObjectIDFromHex(pollHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}

		// 1️⃣ Load poll
		var poll models.Poll
		if err := db.Collection("polls").
			FindOne(context.Background(), bson.M{"_id": pollOID}).
			Decode(&poll); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}

		// 2️⃣ Load YES votes
		cur, err := db.Collection("votes").Find(
			context.Background(),
			bson.M{
				"pollId":    pollOID,
				"attending": true,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer cur.Close(context.Background())

		var votes []struct {
			UserID primitive.ObjectID `bson:"userId"`
		}
		if err := cur.All(context.Background(), &votes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decode error"})
			return
		}

		if len(votes) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"yesCount": 0,
				"teamA":    []TeamPlayer{},
				"teamB":    []TeamPlayer{},
			})
			return
		}

		// 3️⃣ Fetch users by IDs
		userIDs := make([]primitive.ObjectID, 0, len(votes))
		for _, v := range votes {
			userIDs = append(userIDs, v.UserID)
		}

		userCur, err := db.Collection("users").Find(
			context.Background(),
			bson.M{"_id": bson.M{"$in": userIDs}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
			return
		}
		defer userCur.Close(context.Background())

		var users []models.User
		if err := userCur.All(context.Background(), &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user decode error"})
			return
		}

		// 4️⃣ Convert users → team players
		players := make([]TeamPlayer, 0, len(users))
		for _, u := range users {
			players = append(players, TeamPlayer{
				UserID:    u.UserID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Position:  u.Position,
				Skills:    u.Skills,
			})
		}

		// 5️⃣ Shuffle for fairness (skills used later)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		teamA := []TeamPlayer{}
		teamB := []TeamPlayer{}

		for i, p := range players {
			if i%2 == 0 {
				teamA = append(teamA, p)
			} else {
				teamB = append(teamB, p)
			}
		}

		// 6️⃣ Respond
		c.JSON(http.StatusOK, gin.H{
			"yesCount":    len(players),
			"teamA":       teamA,
			"teamB":       teamB,
			"generatedAt": time.Now(),
			"pollDate":    poll.PollDate,
			"pollEndsAt":  poll.EndsAt,
			"pollStatus":  poll.Status,
		})
	}
}
