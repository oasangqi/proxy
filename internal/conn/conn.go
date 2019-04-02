package conn

import (
	"net"

	"github.com/golang/glog"
)

type Conn interface {
	ReadMessage() (payload []byte, err error)
	WriteMessage(payload []byte) (err error)
	RemoteAddr() string
	Close()
	DisConnect()
}

func Transport(from Conn, to Conn, debug bool, whitelist map[string]struct{}) (err error) {
	var (
		buf   []byte
		fAddr = from.RemoteAddr()
		tAddr = to.RemoteAddr()
	)
	for {
		if buf, err = from.ReadMessage(); err != nil {
			to.DisConnect()
			glog.Errorf("read from %s failed, err(%v)", fAddr, err)
			return
		}
		if err = to.WriteMessage(buf); err != nil {
			from.DisConnect()
			glog.Errorf("write to %s failed, err(%v)", tAddr, err)
			return
		}
		if debug {
			ip, _, _ := net.SplitHostPort(fAddr)
			if _, ok := whitelist[ip]; ok {
				if _, ok := from.(*WSConn); ok {
					xorfunc(buf)
					glog.Infof("%s:ws->%s:tcp(%s)", fAddr, tAddr, buf)
				} else {
					glog.Infof("%s:tcp->%s:ws(%s)", fAddr, tAddr, buf)
				}
			}
		}
	}
	return
}

func xorfunc(buf []byte) {
	for i := range buf {
		buf[i] ^= 0x10
	}
}
