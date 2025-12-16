package handlers

import (
	"context"
	"net/http"

	"soccer-app/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoginUser verifies a registered user
func LoginUser(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Secret    string `json:"secret"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if req.FirstName == "" || req.LastName == "" || req.Secret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
			return
		}

		var user models.User
		err := db.Collection("users").FindOne(
			context.Background(),
			bson.M{
				"firstName": req.FirstName,
				"lastName":  req.LastName,
			},
		).Decode(&user)

		if err == mongo.ErrNoDocuments {
			// ❌ Not registered
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		// 🔐 Verify secret
		if user.SecretHash != HashSecret(req.Secret) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid secret"})
			return
		}

		// ✅ Login success
		c.JSON(http.StatusOK, gin.H{
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"position":  user.Position,
			"skills":    user.Skills,
		})
	}
}
