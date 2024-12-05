package network

import "net"

type ConnWithBuffer struct {
	net.Conn
	buff []byte
}

func (c *ConnWithBuffer) Read(b []byte) (n int, err error) {
	if len(c.buff) == 0 {
		return c.Conn.Read(b)
	}

	n = copy(b, c.buff)
	c.buff = (c.buff)[n:]
	return n, nil
}
