package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"soccer-app/config"
	"soccer-app/handlers"
	geo "soccer-app/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type LoginEvent struct {
	FirstName string    `json:"firstName,omitempty"`
	LastName  string    `json:"lastName,omitempty"`
	Time      time.Time `json:"time"`
	UserID    string    `json:"userId,omitempty"`
	IP        string    `json:"ip"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	UserAgent string    `json:"userAgent"`
	Referer   string    `json:"referer,omitempty"`
	Country   string    `json:"country,omitempty"`
	City      string    `json:"city,omitempty"`
	ISP       string    `json:"isp,omitempty"`
	Lat       string    `json:"lat,omitempty"`
	Lng       string    `json:"lng,omitempty"`
}

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
		"ngrok-skip-browser-warning",
	}
	corsCfg.AllowCredentials = true
	corsCfg.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	r.Use(cors.New(corsCfg))

	r.Use(RateLimitMiddleware(100)) // 100 requests per minute

	// 100 requests per minute
	r.SetTrustedProxies(nil) // trust all (OK for dev)

	// ✅ Serve static files
	r.Static("/static", "./static")

	// Pages
	r.GET("/teams", func(c *gin.Context) {
		c.File("./static/teams.html")
	})

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
		api.GET("/polls/:id/teams", handlers.GetTeams(db))
		api.POST("/polls/:id/teams/move", handlers.MovePlayer(db))

		// auth & voting
		api.POST("/register", handlers.RegisterUser(db))
		api.POST("/votes", handlers.SubmitVote(db))

		api.POST("/login", handlers.LoginUser(db))

	}

	log.Println("Server running on :8080")
	log.Fatal(r.Run(":8080"))
}

type RegisterBody struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func WhoLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip CORS preflight
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		ip := c.ClientIP()
		ua := c.GetHeader("User-Agent")
		ref := c.GetHeader("Referer")

		fmt.Println("Client IP:", ip)

		userID := ""
		if uid, ok := c.Get("userId"); ok {
			userID, _ = uid.(string)
		}

		firstName, lastName := "", ""

		// Only parse body where it actually exists
		if c.FullPath() == "/api/v1/register" {
			var rb RegisterBody
			if err := c.ShouldBindBodyWith(&rb, binding.JSON); err == nil {
				firstName = rb.FirstName
				lastName = rb.LastName
			}
		}

		event := LoginEvent{
			Time:      time.Now().UTC(),
			UserID:    userID,
			FirstName: firstName,
			LastName:  lastName,
			IP:        ip,
			Method:    c.Request.Method,
			Path:      c.FullPath(),
			UserAgent: ua,
			Referer:   ref,
		}

		// Geo only for public IPs
		if !isPrivateIP(ip) {
			if geoInfo, err := geo.LookupGeo(ip); err == nil {
				event.Country = geoInfo.Country
				event.City = geoInfo.City
				event.ISP = geoInfo.ISP
				event.Lat = geoInfo.Lat
				event.Lng = geoInfo.Lng
			}
		}

		logLoginEvent(event)
		c.Next()
	}
}

func logLoginEvent(e LoginEvent) {
	// In production → write to DB
	// or structured logger (zap / logrus)
	println(
		e.Method,
		e.FirstName,
		e.LastName,
		"logged in at",
		e.Time.Format(time.RFC3339),
		e.UserID,
		e.IP,
		e.Path,
		e.UserAgent,
		e.Referer,
		"Country:",
		e.Country,
		"CITY:",
		e.City,
		"ISP:",
		e.ISP,
		"LAT:",
		e.Lat,
		"LNG:",
		e.Lng,
	)
}

// RateLimitMiddleware limits the number of requests per minute
func RateLimitMiddleware(maxRequestsPerMinute int) gin.HandlerFunc {
	visitors := make(map[string]int)
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		visitors[clientIP]++
		if visitors[clientIP] > maxRequestsPerMinute {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	return parsed.IsPrivate() || parsed.IsLoopback()
}
