package main

import (
	"Websocket/operate"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	r.GET("/chatroom/:id", operate.Chat)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
