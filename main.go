package main

import (
	"log"
	"net"

	"github.com/OnnaSoft/replica/handler"
	"github.com/OnnaSoft/replica/network"
	"github.com/gin-gonic/gin"
)

func main() {
	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	gin.SetMode(gin.ReleaseMode)
	l := network.NewListener(conn)
	handler := handler.NewHandler()

	go func() {
		for {
			client, err := l.AcceptTCP()
			if err != nil {
				log.Fatal(err)
			}
			go handler.Handle(client)
		}
	}()

	log.Println(handler.RunListener(l))
}
