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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SubmitVote(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID    primitive.ObjectID `json:"userId"`
			PollID    primitive.ObjectID `json:"pollId"`
			FirstName string             `json:"firstName"`
			LastName  string             `json:"lastName"`
			Secret    string             `json:"secret"`
			Rating    int                `json:"rating"`
			Attending bool               `json:"attending"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 🔐 Verify user + secret
		var user models.User
		err := db.Collection("users").FindOne(
			context.Background(),
			bson.M{
				"firstName":  req.FirstName,
				"lastName":   req.LastName,
				"secretHash": HashSecret(req.Secret),
			},
		).Decode(&user)

		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		// 🔁 Upsert vote + ADD userId
		filter := bson.M{
			"pollId": req.PollID,
			"userId": user.UserID, // 🔑 ensures one vote per user per poll
		}

		update := bson.M{
			"$set": bson.M{
				"pollId":    req.PollID,
				"userId":    user.UserID, // ✅ NEW FIELD
				"firstName": req.FirstName,
				"lastName":  req.LastName,
				"rating":    req.Rating,
				"attending": req.Attending,
				"updatedAt": time.Now(),
			},
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
		}

		opts := options.Update().SetUpsert(true)

		if _, err := db.Collection("votes").UpdateOne(
			context.Background(),
			filter,
			update,
			opts,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "vote failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
