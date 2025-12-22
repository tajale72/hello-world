package handlers

import (
	"context"
	"net/http"
	"soccer-app/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func MovePlayer(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// 1️⃣ Parse poll ID
		pollHex := c.Param("id")
		pollOID, err := primitive.ObjectIDFromHex(pollHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}

		// 2️⃣ Parse request
		var req struct {
			UserID   string `json:"userId"`
			FromTeam string `json:"fromTeam"` // "A" or "B"
			ToTeam   string `json:"toTeam"`   // "A" or "B"
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		userOID, err := primitive.ObjectIDFromHex(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		if req.FromTeam == req.ToTeam {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fromTeam and toTeam cannot be same"})
			return
		}

		// 3️⃣ Load existing teams
		var teams models.PollTeams
		err = db.Collection("teams").FindOne(ctx, bson.M{
			"pollId": pollOID,
		}).Decode(&teams)

		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "teams not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		// 4️⃣ Remove player from source team
		var player *models.User

		if req.FromTeam == "A" {
			player, teams.TeamA = extractPlayer(teams.TeamA, userOID)
		} else {
			player, teams.TeamB = extractPlayer(teams.TeamB, userOID)
		}

		if player == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "player not in source team"})
			return
		}

		// 5️⃣ Add to destination team
		if req.ToTeam == "A" {
			teams.TeamA = append(teams.TeamA, *player)
		} else {
			teams.TeamB = append(teams.TeamB, *player)
		}

		teams.UpdatedAt = time.Now()

		// 6️⃣ Persist update (atomic document replace)
		_, err = db.Collection("teams").UpdateOne(
			ctx,
			bson.M{"pollId": pollOID},
			bson.M{
				"$set": bson.M{
					"teamA":     teams.TeamA,
					"teamB":     teams.TeamB,
					"updatedAt": teams.UpdatedAt,
				},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update teams"})
			return
		}

		// 7️⃣ Success
		c.Status(http.StatusOK)
	}
}

func extractPlayer(players []models.User, userID primitive.ObjectID) (*models.User, []models.User) {
	for i, p := range players {
		if p.UserID == userID {
			// remove from slice
			return &p, append(players[:i], players[i+1:]...)
		}
	}
	return nil, players
}
