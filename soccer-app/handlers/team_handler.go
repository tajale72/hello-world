package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"soccer-app/models"
)

type PlayerRating struct {
	Name   string  `json:"name"`
	Rating float64 `json:"rating"`
}

type voteDoc struct {
	FirstName string             `bson:"firstName"`
	LastName  string             `bson:"lastName"`
	Rating    int                `bson:"rating"`
	Attending bool               `bson:"attending"`
	PollID    primitive.ObjectID `bson:"pollId"`
}

func GenerateTeams(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Parse poll id
		pollHex := c.Param("id")
		pollOID, err := primitive.ObjectIDFromHex(pollHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}

		// ✅ 2) Load poll so `poll` is defined (fixes your compile error)
		var poll models.Poll
		if err := db.Collection("polls").FindOne(context.Background(), bson.M{"_id": pollOID}).Decode(&poll); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}

		// (Optional) if you want to block team generation after close:
		// if time.Now().After(poll.EndsAt) || poll.Status != "OPEN" {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "poll closed"})
		// 	return
		// }

		// 3) Fetch ALL YES votes for this poll
		cur, err := db.Collection("votes").Find(context.Background(), bson.M{
			"pollId":    pollOID,
			"attending": true,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer cur.Close(context.Background())

		votes := make([]voteDoc, 0)
		if err := cur.All(context.Background(), &votes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decode error"})
			return
		}

		// 4) Build player list
		players := make([]PlayerRating, 0, len(votes))
		for _, v := range votes {
			full := strings.TrimSpace(v.FirstName + " " + v.LastName)
			if full == "" {
				full = "Anonymous"
			}
			players = append(players, PlayerRating{
				Name:   full,
				Rating: float64(v.Rating),
			})
		}

		// 5) Sort and split into teams
		sort.Slice(players, func(i, j int) bool {
			return players[i].Rating > players[j].Rating
		})

		teamA := make([]PlayerRating, 0)
		teamB := make([]PlayerRating, 0)

		for i, p := range players {
			if i%2 == 0 {
				teamA = append(teamA, p)
			} else {
				teamB = append(teamB, p)
			}
		}

		// 6) Return teams + YES count + timestamps + poll deadline info
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
