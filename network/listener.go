package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
)

// IncomingConn wraps a network connection and an optional error.
type IncomingConn struct {
	net.Conn
	Err error
}

// Listener wraps a net.Listener and provides separate channels for HTTP and TCP connections.
type Listener struct {
	net.Listener
	incomingHTTP chan *IncomingConn
	incomingTCP  chan *IncomingConn
}

// NewListener initializes a Listener and starts proxying connections.
func NewListener(l net.Listener) *Listener {
	nl := &Listener{
		Listener:     l,
		incomingHTTP: make(chan *IncomingConn),
		incomingTCP:  make(chan *IncomingConn),
	}
	go nl.proxy()
	return nl
}

// proxy accepts connections and determines whether they are HTTP or TCP.
func (l *Listener) proxy() {
	buffer := make([]byte, 1024) // Allocate buffer once for efficiency.
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			// Handle listener-level errors.
			l.sendError(err)
			return
		}

		fmt.Println("Accepted connection")

		// Read from the connection to identify the type.
		n, err := conn.Read(buffer)
		if err != nil {
			// Handle connection-level read errors.
			l.sendError(err)
			conn.Close()
			continue
		}

		lconn := &ConnWithBuffer{conn, buffer[:n]}
		reader := bufio.NewReader(bytes.NewBuffer(buffer[:n]))

		// Attempt to parse the data as an HTTP request.
		if _, err := http.ReadRequest(reader); err == nil {
			l.incomingHTTP <- &IncomingConn{lconn, nil}
		} else {
			l.incomingTCP <- &IncomingConn{lconn, nil}
		}
	}
}

// AcceptTCP accepts a TCP connection from the incomingTCP channel.
func (l *Listener) AcceptTCP() (net.Conn, error) {
	fmt.Println("Accepting TCP connection")
	incoming := <-l.incomingTCP
	return incoming.Conn, incoming.Err
}

// Accept accepts an HTTP connection from the incomingHTTP channel.
func (l *Listener) AcceptHTTP() (net.Conn, error) {
	fmt.Println("Accepting HTTP connection")
	incoming := <-l.incomingHTTP
	return incoming.Conn, incoming.Err
}

func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptHTTP()
}

// Close closes the Listener and associated channels.
func (l *Listener) Close() error {
	err := fmt.Errorf("listener closed")
	// Notify all goroutines waiting for connections.
	l.sendError(err)
	close(l.incomingHTTP)
	close(l.incomingTCP)
	return l.Listener.Close()
}

// Addr returns the address of the underlying Listener.
func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}

// sendError sends the given error to both incomingHTTP and incomingTCP channels.
func (l *Listener) sendError(err error) {
	l.incomingHTTP <- &IncomingConn{nil, err}
	l.incomingTCP <- &IncomingConn{nil, err}
}
