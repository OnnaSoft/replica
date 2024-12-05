package network

import "net"

// ConnWithBuffer wraps a net.Conn and includes a read buffer.
type ConnWithBuffer struct {
	net.Conn
	Buffer []byte
}

func (c *ConnWithBuffer) Read(b []byte) (n int, err error) {
	if len(c.Buffer) == 0 {
		return c.Conn.Read(b)
	}

	n = copy(b, c.Buffer)
	c.Buffer = (c.Buffer)[n:]
	return n, nil
}
