package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
)

type IncomingConn struct {
	net.Conn
	error
}

type Listener struct {
	net.Listener
	incomingHTTP chan *IncomingConn
	incomingTCP  chan *IncomingConn
}

func NewListener(l net.Listener) *Listener {
	nl := &Listener{
		Listener:     l,
		incomingHTTP: make(chan *IncomingConn),
		incomingTCP:  make(chan *IncomingConn),
	}
	go nl.proxy()

	return nl
}

func (l *Listener) proxy() {
	b := make([]byte, 1024)
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			l.incomingHTTP <- &IncomingConn{nil, err}
			return
		}

		n, err := conn.Read(b)
		if err != nil {
			l.incomingHTTP <- &IncomingConn{nil, err}
			return
		}
		lconn := &ConnWithBuffer{conn, b[:n]}

		reader := bufio.NewReader(bytes.NewBuffer(b[:n]))
		if _, err := http.ReadRequest(reader); err == nil {
			l.incomingHTTP <- &IncomingConn{lconn, nil}
			continue
		}

		l.incomingTCP <- &IncomingConn{lconn, nil}
	}
}

func (l *Listener) AcceptTCP() (conn net.Conn, err error) {
	incoming := <-l.incomingTCP
	return incoming.Conn, incoming.error
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	incoming := <-l.incomingHTTP
	return incoming.Conn, incoming.error
}

func (l *Listener) Close() error {
	l.incomingHTTP <- &IncomingConn{nil, fmt.Errorf("listener closed")}
	l.incomingTCP <- &IncomingConn{nil, fmt.Errorf("listener closed")}
	close(l.incomingHTTP)
	close(l.incomingTCP)
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}
