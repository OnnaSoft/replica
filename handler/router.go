package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

func (h *Handler) setupRouter() {
	r := gin.Default()

	r.POST("/publish", func(c *gin.Context) {
		var msg Message
		if err := c.BindJSON(&msg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("Publishing |" + msg.Data + "| to |" + msg.Channel + "|")
		go h.Publish(msg.Channel, msg.Data)

		fmt.Println("Published |" + msg.Data + "| to |" + msg.Channel + "|")

		c.JSON(200, gin.H{"status": "ok"})
	})

	h.Engine = r
}
