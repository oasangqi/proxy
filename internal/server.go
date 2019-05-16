package proxy

import (
	"github.com/oasangqi/proxy/internal/conf"
	"github.com/oasangqi/proxy/internal/dao"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	maxInt = 1<<31 - 1
)

type Server struct {
	cfg             *conf.Config
	MonitorInterval time.Duration
	round           *Round
	dao             *dao.Dao
	donec           chan struct{}
	mu              sync.RWMutex // guards the following fields
	Online          map[int64]*Online
	WhiteList       map[string]struct{}
}

func New(c *conf.Config) (*Server, error) {
	if c.Interval == 0 {
		glog.Infof("invalid interval:%d, set to 5", c.Interval)
		c.Interval = 5
	}

	return &Server{
		cfg:             c,
		MonitorInterval: time.Duration(c.Interval),
		round:           NewRound(c),
		dao:             dao.New(c),
		donec:           make(chan struct{}),
		Online:          newOnLine(c),
		WhiteList:       make(map[string]struct{}),
	}, nil
}

func (srv *Server) Run() {
	glog.Infof("proxy start running")
	go srv.runMonitor()
	go srv.initWebSocket()
	if srv.cfg.WebSocket.TlsOpen {
		go srv.initWebSocketTLS()
	}
}

func (srv *Server) Close() {
}

func (srv *Server) runMonitor() {
	for {
		glog.Flush()
		time.Sleep(srv.MonitorInterval)
	}
}
