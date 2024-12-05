package handler

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	subscribers sync.Map
}

func NewHandler() *Handler {
	h := &Handler{
		subscribers: sync.Map{},
	}
	h.setupRouter()

	return h
}

func (h *Handler) Subscribe(id string, c net.Conn) {
	fmt.Println("Subscribing to |" + id + "|")
	h.subscribers.Store(id, c)
}

func (h *Handler) Publish(id, message string) {
	if v, ok := h.subscribers.Load(id); ok {
		c := v.(net.Conn)
		fmt.Println("Publishing |" + message + "|")
		c.Write([]byte(message + "\n"))
	}
}

func (h *Handler) Unsubscribe(id string) {
	fmt.Println("Unsubscribing from |" + id + "|")
	h.subscribers.Delete(id)
}

func (h *Handler) Handle(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	channels := make([]string, 0)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		channel := strings.Trim(line[:len(line)-1], "\r")
		fmt.Println("Received |" + channel + "|")
		channels = append(channels, channel)
		h.Subscribe(channel, c)
		fmt.Println("Subscribed to |" + channel + "|")
	}
}
