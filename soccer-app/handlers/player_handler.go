package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"soccer-app/models"
)

func GetPlayers(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var players []models.Player
		cur, _ := db.Collection("players").Find(context.Background(), map[string]interface{}{})
		cur.All(context.Background(), &players)
		c.JSON(http.StatusOK, players)
	}
}

func CreatePlayer(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p models.Player
		c.ShouldBindJSON(&p)
		db.Collection("players").InsertOne(context.Background(), p)
		c.JSON(http.StatusCreated, p)
	}
}