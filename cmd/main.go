package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jayanthdeejay/trending/internal"
)

func main() {
	// internal.FetchAndSaveFuturesData()
	go internal.ConnectToWebSocket(internal.UsdtSymbols)

	// Set the router as the default one shipped with Gin
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Binance Futures Tracker",
		})
	})
	log.Println("Starting Gin server on 0.0.0.0:8080")
	if err := r.Run(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
