package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Binance Futures Tracker",
		})
	})
	r.Run() // Listen and serve on 0.0.0.0:8080
}
