package handler

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/gin-gonic/gin"
)

// Handler struct manages the Gin engine and subscribers for publish/subscribe functionality.
type Handler struct {
	*gin.Engine
	subscribers sync.Map // Concurrent map to store subscribers.
}

// NewHandler initializes a new Handler instance.
func NewHandler() *Handler {
	h := &Handler{
		subscribers: sync.Map{},
	}
	h.setupRouter() // Assuming setupRouter configures Gin routes (not shown here).
	return h
}

// Subscribe adds a connection to the subscribers map for a given ID.
func (h *Handler) Subscribe(id string, c net.Conn) {
	fmt.Printf("Subscribing to |%s|\n", id)
	if v, ok := h.subscribers.Load(id); ok {
		conns := v.([]net.Conn)
		conns = append(conns, c)
		h.subscribers.Store(id, conns)
		return
	}

	h.subscribers.Store(id, []net.Conn{c})
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Publish sends a message to the subscriber identified by ID.
func (h *Handler) Publish(id, message string) {
	if v, ok := h.subscribers.Load(id); ok {
		conns := v.([]net.Conn)
		conn := conns[rnd.Intn(len(conns))]
		fmt.Printf("Publishing |%s| to |%s|\n", message, id)
		_, err := conn.Write([]byte(message + "\n"))
		if err != nil {
			log.Printf("Error publishing to %s: %v\n", id, err)
		}
	} else {
		fmt.Printf("No subscriber found for ID: |%s|\n", id)
	}
}

// Unsubscribe removes a subscriber by ID.
func (h *Handler) Unsubscribe(id string) {
	fmt.Printf("Unsubscribing from |%s|\n", id)
	h.subscribers.Delete(id)
}

// Handle manages a client connection, subscribing it to channels based on input.
func (h *Handler) Handle(c net.Conn) {
	defer c.Close()

	reader := bufio.NewReader(c)
	channels := make([]string, 0) // Tracks channels the connection is subscribed to.

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection closed or error: %v\n", err)
			h.cleanupSubscriptions(channels, c)
			return
		}

		// Clean up the received channel ID.
		channel := strings.TrimSpace(line)
		fmt.Printf("Received |%s|\n", channel)

		// Prevent duplicate subscriptions to the same channel.
		if !contains(channels, channel) {
			channels = append(channels, channel)
			h.Subscribe(channel, c)
			fmt.Printf("Subscribed to |%s|\n", channel)
		} else {
			fmt.Printf("Already subscribed to |%s|\n", channel)
		}
	}
}

// cleanupSubscriptions removes the connection from all subscribed channels.
func (h *Handler) cleanupSubscriptions(channels []string, c net.Conn) {
	for _, channel := range channels {
		if v, ok := h.subscribers.Load(channel); ok && v == c {
			h.Unsubscribe(channel)
		}
	}
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
