package main

import (
	"log"

	"soccer-app/config"
	"soccer-app/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	db := config.MustMongo()

	r := gin.Default()

	// ✅ Serve static files
	r.Static("/static", "./static")

	// ✅ Serve index.html at root "/"
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	api := r.Group("/api/v1")
	{
		api.GET("/players", handlers.GetPlayers(db))
		api.POST("/players", handlers.CreatePlayer(db))

		api.POST("/polls", handlers.CreatePoll(db))
		api.GET("/polls/current", handlers.GetCurrentPoll(db))

		api.POST("/votes", handlers.SubmitVote(db))
		api.POST("/polls/:id/teams", handlers.GenerateTeams(db))
	}

	log.Println("Server running on :8080")
	r.Run(":8080")
}
