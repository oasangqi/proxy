package conn

import (
	"encoding/binary"
	"errors"
	"net"
	"proxy/pkg/bufio"
	"proxy/pkg/bytes"
)

type TCPConn struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
	rb   *bytes.Buffer
	wb   *bytes.Buffer
	rp   *bytes.Pool
	wp   *bytes.Pool
}

func NewTcpConn(conn net.Conn, rp, wp *bytes.Pool) *TCPConn {
	tc := &TCPConn{
		conn: conn,
		r:    new(bufio.Reader),
		w:    new(bufio.Writer),
		rp:   rp,
		wp:   wp,
		rb:   rp.Get(),
		wb:   wp.Get(),
	}
	tc.r.ResetBuffer(conn, tc.rb.Bytes())
	tc.w.ResetBuffer(conn, tc.wb.Bytes())
	return tc
}

func (c *TCPConn) ReadMessage() (payload []byte, err error) {
	var (
		packLen uint32
		buf     []byte
	)
	if buf, err = c.r.Pop(4); err != nil {
		return
	}
	packLen = binary.LittleEndian.Uint32(buf[0:4])
	if packLen > 16*1024 {
		return nil, errors.New("invalid pack size")
	}
	if packLen > 0 {
		payload, err = c.r.Pop(int(packLen))
	}
	xorfunc(payload)
	return
}

func (c *TCPConn) WriteMessage(payload []byte) (err error) {
	var (
		buf     []byte
		packLen uint32
	)
	packLen = uint32(len(payload))
	if buf, err = c.w.Peek(4); err != nil {
		return
	}
	binary.LittleEndian.PutUint32(buf[0:4], packLen)
	if packLen > 0 {
		xorfunc(payload)
		_, err = c.w.Write(payload)
		c.w.Flush()
	}
	return
}

func (c *TCPConn) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *TCPConn) Close() {
	c.conn.Close()
	c.rp.Put(c.rb)
	c.wp.Put(c.wb)
}

func (c *TCPConn) DisConnect() {
	c.conn.Close()
}
