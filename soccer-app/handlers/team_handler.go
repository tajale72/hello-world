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
		ctx := context.Background()

		pollHex := c.Param("id")
		pollOID, err := primitive.ObjectIDFromHex(pollHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}

		// 1️⃣ Load poll
		var poll models.Poll
		if err := db.Collection("polls").
			FindOne(ctx, bson.M{"_id": pollOID}).
			Decode(&poll); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}

		// 2️⃣ CHECK IF TEAMS ALREADY EXIST (CRITICAL)
		var existing models.PollTeams
		err = db.Collection("teams").
			FindOne(ctx, bson.M{"pollId": pollOID}).
			Decode(&existing)

		if err == nil {
			// ✅ Teams already generated → just return DB state
			c.JSON(http.StatusOK, gin.H{
				"yesCount":    existing.YesCount,
				"teamA":       existing.TeamA,
				"teamB":       existing.TeamB,
				"generatedAt": existing.GeneratedAt,
				"pollDate":    poll.PollDate,
				"pollEndsAt":  poll.EndsAt,
				"pollStatus":  poll.Status,
			})
			return
		}

		// 3️⃣ Load YES votes
		cur, err := db.Collection("votes").Find(ctx, bson.M{
			"pollId":    pollOID,
			"attending": true,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		defer cur.Close(ctx)

		var votes []struct {
			UserID primitive.ObjectID `bson:"userId"`
		}
		if err := cur.All(ctx, &votes); err != nil {
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

		// 4️⃣ Load users
		userIDs := make([]primitive.ObjectID, 0, len(votes))
		for _, v := range votes {
			userIDs = append(userIDs, v.UserID)
		}

		userCur, err := db.Collection("users").Find(ctx, bson.M{
			"_id": bson.M{"$in": userIDs},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
			return
		}
		defer userCur.Close(ctx)

		var users []models.User
		if err := userCur.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user decode error"})
			return
		}

		// 5️⃣ Convert to team players
		players := make([]models.User, 0, len(users))
		for _, u := range users {
			players = append(players, models.User{
				UserID:    u.UserID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Position:  u.Position,
				Skills:    u.Skills,
			})
		}

		// 6️⃣ Shuffle ONCE
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(players), func(i, j int) {
			players[i], players[j] = players[j], players[i]
		})

		var teamA, teamB []models.User
		for i, p := range players {
			if i%2 == 0 {
				teamA = append(teamA, p)
			} else {
				teamB = append(teamB, p)
			}
		}

		// 7️⃣ PERSIST TEAMS (MOST IMPORTANT STEP)
		teamsDoc := models.PollTeams{
			PollID:      pollOID,
			TeamA:       teamA,
			TeamB:       teamB,
			YesCount:    len(players),
			GeneratedAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err = db.Collection("teams").InsertOne(ctx, teamsDoc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save teams"})
			return
		}

		// 8️⃣ Return DB-backed response
		c.JSON(http.StatusOK, gin.H{
			"yesCount":    teamsDoc.YesCount,
			"teamA":       teamA,
			"teamB":       teamB,
			"generatedAt": teamsDoc.GeneratedAt,
			"pollDate":    poll.PollDate,
			"pollEndsAt":  poll.EndsAt,
			"pollStatus":  poll.Status,
		})
	}
}

func GetTeams(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// 1️⃣ Parse poll ID
		pollHex := c.Param("id")
		pollOID, err := primitive.ObjectIDFromHex(pollHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
			return
		}

		// 2️⃣ Load poll (for metadata only)
		var poll models.Poll
		if err := db.Collection("polls").
			FindOne(ctx, bson.M{"_id": pollOID}).
			Decode(&poll); err != nil {

			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}

		// 3️⃣ Load teams from TEAMS collection (SOURCE OF TRUTH)
		var teams models.PollTeams
		err = db.Collection("teams").
			FindOne(ctx, bson.M{"pollId": pollOID}).
			Decode(&teams)

		if err == mongo.ErrNoDocuments {
			// Teams not generated yet
			c.JSON(http.StatusOK, gin.H{
				"yesCount":   0,
				"teamA":      []models.User{},
				"teamB":      []models.User{},
				"pollDate":   poll.PollDate,
				"pollEndsAt": poll.EndsAt,
				"pollStatus": poll.Status,
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load teams"})
			return
		}

		// 4️⃣ Return DB-backed response
		c.JSON(http.StatusOK, gin.H{
			"yesCount":    teams.YesCount,
			"teamA":       teams.TeamA,
			"teamB":       teams.TeamB,
			"generatedAt": teams.GeneratedAt,
			"updatedAt":   teams.UpdatedAt,
			"pollDate":    poll.PollDate,
			"pollEndsAt":  poll.EndsAt,
			"pollStatus":  poll.Status,
		})
	}
}
