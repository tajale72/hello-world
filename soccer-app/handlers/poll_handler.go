package handlers

import (
	"context"
	"net/http"
	"time"

	"soccer-app/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type createPollReq struct {
	// Optional. If empty, defaults to next Saturday.
	PollDate string `json:"pollDate"`
	// Optional. If empty, defaults to Saturday 10:00 AM America/Chicago.
	EndsAt string `json:"endsAt"` // RFC3339 recommended
}

func nextSaturdayDate(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	// Go: Sunday=0 ... Saturday=6
	days := (6 - int(now.Weekday()) + 7) % 7
	if days == 0 {
		days = 7 // if today is Saturday, use next Saturday
	}
	d := now.AddDate(0, 0, days)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
}

func CreatePoll(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc, _ := time.LoadLocation("America/Chicago")

		var req createPollReq
		_ = c.ShouldBindJSON(&req)

		// pollDate
		var pollDay time.Time
		if req.PollDate != "" {
			// expects YYYY-MM-DD
			t, err := time.ParseInLocation("2006-01-02", req.PollDate, loc)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pollDate, use YYYY-MM-DD"})
				return
			}
			pollDay = t
		} else {
			pollDay = nextSaturdayDate(loc)
		}

		// endsAt
		var endsAt time.Time
		if req.EndsAt != "" {
			t, err := time.Parse(time.RFC3339, req.EndsAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endsAt, use RFC3339 مثل 2025-12-13T10:00:00-06:00"})
				return
			}
			endsAt = t.In(loc)
		} else {
			// default: Saturday 10:00 AM local time
			endsAt = time.Date(pollDay.Year(), pollDay.Month(), pollDay.Day(), 10, 0, 0, 0, loc)
		}

		poll := models.Poll{
			PollDate:  pollDay.Format("2006-01-02"),
			Status:    "OPEN",
			EndsAt:    endsAt,
			CreatedAt: time.Now().In(loc),
		}

		_, err := db.Collection("polls").InsertOne(context.Background(), poll)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
		}

		c.JSON(http.StatusCreated, poll)
	}
}

func GetCurrentPoll(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc, _ := time.LoadLocation("America/Chicago")
		now := time.Now().In(loc)

		var poll models.Poll
		err := db.Collection("polls").FindOne(
			context.Background(),
			bson.M{
				"status": "OPEN",
				"endsAt": bson.M{"$gt": now}, // only not-expired poll
			},
		).Decode(&poll)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no open poll"})
			return
		}

		c.JSON(http.StatusOK, poll)
	}
}
