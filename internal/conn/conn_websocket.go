package conn

import (
	"errors"
	"fmt"
	"net"
	"proxy/pkg/bufio"
	"proxy/pkg/bytes"
	"proxy/pkg/websocket"

	"github.com/golang/glog"
)

type WSConn struct {
	ws *websocket.Conn
	rb *bytes.Buffer
	wb *bytes.Buffer
	rp *bytes.Pool
	wp *bytes.Pool
}

func NewWSConn(conn net.Conn, rp, wp *bytes.Pool) *WSConn {
	wc := &WSConn{
		rp: rp,
		wp: wp,
		rb: rp.Get(),
		wb: wp.Get(),
	}
	r := new(bufio.Reader)
	w := new(bufio.Writer)
	r.ResetBuffer(conn, wc.rb.Bytes())
	w.ResetBuffer(conn, wc.wb.Bytes())
	wc.ws = websocket.NewConn(conn, r, w)
	return wc
}

func (c *WSConn) ReadMessage() (payload []byte, err error) {

	if _, payload, err = c.ws.ReadMessage(); err != nil {
		glog.Errorf("ws read failed, len:%d payload(%s) error(%v)", len(payload), payload, err)
		return
	}
	return
}

func (c *WSConn) WriteMessage(payload []byte) (err error) {
	var (
		packLen int
	)
	packLen = len(payload)
	if err = c.ws.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		glog.Errorf("ws write head failed, err(%v)", err)
		return
	}
	if packLen > 0 {
		err = c.ws.WriteBody(payload)
	}
	c.ws.Flush()
	return
}

func (c *WSConn) HandShake() (err error) {
	var (
		req *websocket.Request
	)
	if req, err = websocket.ReadRequest(c.ws); err != nil {
		glog.Errorf("ws read req failed error(%v)", err)
		return
	}
	if req.RequestURI != "/ws" {
		err = errors.New(fmt.Sprintf("invalid URI:%s", req.RequestURI))
		glog.Errorf("ws read req failed error(%v)", err)
		return
	}
	if err = websocket.Upgrade(c.ws, req); err != nil {
		glog.Errorf("ws upgrate failed error(%v)", err)
		return
	}
	return
}

// TODO: c.ws已实现了该方法，为何c实现接口Conn时还需要实现该方法
func (c *WSConn) RemoteAddr() string {
	return c.ws.RemoteAddr()
}

func (c *WSConn) Close() {
	c.ws.Close()
	c.rp.Put(c.rb)
	c.wp.Put(c.wb)
}

func (c *WSConn) DisConnect() {
	c.ws.Close()
}
