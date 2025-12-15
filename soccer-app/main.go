package main

import (
	"log"

	"soccer-app/config"
	"soccer-app/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.MustMongo()

	r := gin.Default()

	// CORS middleware
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsCfg.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
	}
	corsCfg.AllowCredentials = true
	corsCfg.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	r.Use(cors.New(corsCfg))

	// ✅ Serve static files
	r.Static("/static", "./static")

	// Pages
	r.GET("/", func(c *gin.Context) {
		c.File("./static/signin.html")
	})

	r.GET("/register", func(c *gin.Context) {
		c.File("./static/register.html")
	})

	r.GET("/game", func(c *gin.Context) {
		c.File("./static/game.html")
	})

	r.GET("/index", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// ✅ API v1 routes (REGISTER ONCE)
	api := r.Group("/api/v1")
	{
		// players
		api.GET("/players", handlers.GetPlayers(db))
		api.POST("/players", handlers.CreatePlayer(db))

		// polls
		api.POST("/polls", handlers.CreatePoll(db))
		api.GET("/polls/current", handlers.GetCurrentPoll(db))
		api.POST("/polls/:id/teams", handlers.GenerateTeams(db))

		// auth & voting
		api.POST("/register", handlers.RegisterUser(db))
		api.POST("/votes", handlers.SubmitVote(db))
	}

	log.Println("Server running on :8080")
	log.Fatal(r.Run(":8080"))
}
