package handlers

import (
	"context"
	"net/http"
	"time"

	"soccer-app/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SubmitVote(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var v models.Vote
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		var poll models.Poll
		if err := db.Collection("polls").FindOne(context.Background(), bson.M{"_id": v.PollID}).Decode(&poll); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "poll not found"})
			return
		}
		if time.Now().After(poll.EndsAt) || poll.Status != "OPEN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "poll closed"})
			return
		}

		// convert pollId string to ObjectID
		pollOID, err := primitive.ObjectIDFromHex(v.PollID.Hex())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}
		v.PollID = pollOID

		// prevent duplicate YES votes
		if v.Attending {
			count, _ := db.Collection("votes").CountDocuments(context.Background(), bson.M{
				"pollId":    v.PollID,
				"firstName": v.FirstName,
				"lastName":  v.LastName,
				"attending": true,
			})
			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "already voted yes"})
				return
			}
		}

		if err := db.Collection("polls").FindOne(context.Background(), bson.M{"_id": v.PollID}).Decode(&poll); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "poll not found"})
			return
		}
		if time.Now().After(poll.EndsAt) || poll.Status != "OPEN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "poll closed"})
			return
		}

		v.CreatedAt = time.Now()
		if _, err := db.Collection("votes").InsertOne(context.Background(), v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
