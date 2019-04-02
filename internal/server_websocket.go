package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"proxy/internal/conn"
	"runtime"
	"strings"
	"time"

	"github.com/golang/glog"
)

const (
	CMD_LOGIN            = 1001
	REPLY_ROOM_NOT_EXIST = "{\"cmd\":4002,\"code\":500,\"msg\":\"房间不存在\"}"
)

func (srv *Server) initWebSocket() (err error) {
	var (
		bind     string
		addr     *net.TCPAddr
		listener *net.TCPListener
	)
	for _, bind = range srv.cfg.WebSocket.Addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			glog.Errorf("net.ResolveTCPAddr(tcp, %s) failed err(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			glog.Errorf("net.ListenTCP(tcp, %s) failed", addr)
			return
		}
		for i := 0; i < runtime.NumCPU(); i++ {
			go srv.acceptWebSocket(listener)
		}
	}
	return
}

func (srv *Server) acceptWebSocket(l *net.TCPListener) {
	var (
		r int
	)
	for {
		in, err := l.Accept()
		if err != nil {
			glog.Errorf("WebSocket accept failed, err(%v)", err)
			return
		}
		go srv.serveWebSocket(in, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func (srv *Server) initWebSocketTLS() (err error) {
	var (
		listener net.Listener
		cert     tls.Certificate
		certs    []tls.Certificate
	)
	certFiles := strings.Split(srv.cfg.WebSocket.CertFile, ",")
	privateFiles := strings.Split(srv.cfg.WebSocket.PrivateFile, ",")
	for i := range certFiles {
		cert, err = tls.LoadX509KeyPair(certFiles[i], privateFiles[i])
		if err != nil {
			glog.Errorf("Error loading certificate. error(%v)", err)
			return
		}
		certs = append(certs, cert)
	}
	tlsCfg := &tls.Config{Certificates: certs}
	for _, addr := range srv.cfg.WebSocket.TlsAddrs {
		if listener, err = tls.Listen("tcp", addr, tlsCfg); err != nil {
			glog.Errorf("net.ListenTCP(tcp, %s) error(%v)", addr, err)
			return
		}
		glog.Infof("start wss listen: %s", addr)
		for i := 0; i < runtime.NumCPU(); i++ {
			go srv.acceptWebSocketTLS(listener)
		}
	}
	return
}

func (srv *Server) acceptWebSocketTLS(l net.Listener) {
	var (
		r int
	)
	for {
		in, err := l.Accept()
		if err != nil {
			glog.Errorf("TLS WebSocket accept failed, err(%v)", err)
			return
		}
		go srv.serveWebSocket(in, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func (srv *Server) serveWebSocket(in net.Conn, r int) {
	var (
		err    error
		out    net.Conn
		online Online
		auth   []byte
		lAddr  = in.LocalAddr().String()
		rAddr  = in.RemoteAddr().String()
		tr     = srv.round.Timer(r)
		rp     = srv.round.Reader(r)
		wp     = srv.round.Writer(r)
	)
	glog.Infof("start to serve \"%s\" with \"%s\"", lAddr, rAddr)
	// websocket handshake
	from := conn.NewWSConn(in, rp, wp)
	defer from.Close()
	trd := tr.Add(time.Duration(srv.cfg.WebSocket.HandShakeTimeOut), func() {
		from.DisConnect()
		glog.Errorf("ws client:%s handshake timeout:%d", rAddr, srv.cfg.WebSocket.HandShakeTimeOut)
	})
	if err = from.HandShake(); err != nil {
		glog.Errorf("ws client:%s handshake failed error(%v)", rAddr, err)
		tr.Del(trd)
		return
	}
	// get server addr
	if online, auth, err = srv.getOnlineServer(from); err != nil {
		glog.Errorf("get remote game server failed, error(%v)", err)
		tr.Del(trd)
		return
	}
	// connect server
	if out, err = net.DialTimeout("tcp", online.Addr, time.Duration(srv.cfg.WebSocket.HandShakeTimeOut)); err != nil {
		glog.Errorf("ws client:%s connect to (%s) failed error(%v)", rAddr, online.Addr, err)
		tr.Del(trd)
		return
	}
	tr.Del(trd)
	// transport message
	to := conn.NewTcpConn(out, rp, wp)
	defer to.Close()
	to.WriteMessage(auth)
	ip, _, _ := net.SplitHostPort(rAddr)
	srv.incOnLine(online.Vid, ip)
	go conn.Transport(from, to, srv.cfg.WebSocket.Debug, srv.WhiteList)
	conn.Transport(to, from, srv.cfg.WebSocket.Debug, srv.WhiteList)
	srv.descOnLine(online.Vid, ip)
}

func (srv *Server) getOnlineServer(c conn.Conn) (online Online, auth []byte, err error) {
	var (
		vid int64
		ok  bool
		ol  *Online
		js  struct {
			Cmd  int64
			Room int64
			Skey string
			Uid  int64
		}
	)
	if auth, err = c.ReadMessage(); err != nil {
		glog.Errorf("auth failed, read failed")
		return
	}
	err = json.Unmarshal(auth, &js)
	switch {
	case err != nil:
		glog.Errorf("auth failed, invalid json:%s", auth[:])
		return
		/*
			case js.Cmd != CMD_LOGIN:
				err = errors.New(fmt.Sprintf("auth failed, invalid cmd:%d", js.Cmd))
				glog.Errorf("%v", err)
				return
		*/
	}
	if vid, err = srv.dao.GetServerVid(context.TODO(), js.Room); err != nil {
		glog.Errorf("auth failed, get room:%d vid failed", js.Room)
		c.WriteMessage([]byte(REPLY_ROOM_NOT_EXIST))
		return
	}
	if ol, ok = srv.Online[vid]; !ok {
		err = errors.New(fmt.Sprintf("auth failed, vid:%d not support", vid))
		glog.Errorf("%v", err)
		return
	}
	online = *ol
	return
}
