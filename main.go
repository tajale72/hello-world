package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	router := gin.Default()
	router.GET("/", Hello)
	router.Run(port)
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusAccepted, "hello world")
}
