package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"soccer-app/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var allowedSkills = map[string]bool{
	"speed":     true,
	"dribbling": true,
	"shooting":  true,
	"defending": true,
	"passing":   true,
	"stamina":   true,
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func RegisterUser(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req struct {
			FirstName string         `json:"firstName"`
			LastName  string         `json:"lastName"`
			Position  string         `json:"position"`
			Skills    []models.Skill `json:"skills"`
			Secret    string         `json:"secret"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.FirstName == "" || req.LastName == "" || req.Secret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "firstName, lastName and secret are required"})
			return
		}

		if len(req.Skills) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one skill is required"})
			return
		}

		// ✅ Validate skills
		for _, s := range req.Skills {
			name := strings.ToLower(s.Name)

			if !allowedSkills[name] {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid skill: " + s.Name,
				})
				return
			}

			if s.Value < 1 || s.Value > 10 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "skill value must be between 1 and 10",
				})
				return
			}

			s.Name = name
		}

		filter := bson.M{
			"firstName": req.FirstName,
			"lastName":  req.LastName,
		}

		update := bson.M{
			"$setOnInsert": bson.M{
				"firstName":  req.FirstName,
				"lastName":   req.LastName,
				"secretHash": hashSecret(req.Secret),
				"createdAt":  time.Now(),
			},
			"$set": bson.M{
				"position": req.Position,
				"skills":   req.Skills,
			},
		}

		opts := options.Update().SetUpsert(true)

		_, err := db.Collection("users").UpdateOne(
			context.Background(),
			filter,
			update,
			opts,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"success": true})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
