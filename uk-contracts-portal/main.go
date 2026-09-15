package main

import (
	"uk-contracts-portal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Serve index.html on root
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Register API routes
	handlers.RegisterRoutes(router)

	// Start server on port 8080
	router.Run(":8080")
}
