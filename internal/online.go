package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"proxy/internal/conf"
)

type Online struct {
	Vid   int64
	Name  string
	Addr  string
	IPs   map[string]int64
	Total int64
}

func newOnLine(c *conf.Config) map[int64]*Online {
	onlines := make(map[int64]*Online)
	for _, r := range c.Servers {
		onlines[r.Vid] = &Online{
			Vid:   r.Vid,
			Name:  r.Name,
			Addr:  r.Addr,
			IPs:   make(map[string]int64),
			Total: 0,
		}
	}
	return onlines
}

func (srv *Server) incOnLine(vid int64, ip string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if online, ok := srv.Online[vid]; ok {
		online.IPs[ip]++
		online.Total++
	}
}

func (srv *Server) descOnLine(vid int64, ip string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	online, ok := srv.Online[vid]
	if !ok {
		return
	}
	if online.IPs[ip] > 1 {
		online.IPs[ip]--
	} else {
		delete(online.IPs, ip)
	}
	online.Total--
}

func (srv *Server) OnlineVid(c context.Context, vid int64) (online *Online, err error) {
	online, ok := srv.Online[vid]
	if !ok {
		online = &Online{}
	}
	return
}

func (srv *Server) OnlineAll(c context.Context) (onlines []*Online, err error) {

	onlines = []*Online{}
	for _, online := range srv.Online {
		onlines = append(onlines, online)
	}
	return
}

func (srv *Server) AddOnline(c context.Context, msg []byte) (onlines []*Online, err error) {
	var (
		ol Online
	)
	if err = json.Unmarshal(msg, &ol); err != nil {
		return
	}
	if len(ol.Addr) == 0 {
		err = errors.New(fmt.Sprintf("invalid addr:%s", ol.Addr))
		return
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if _, ok := srv.Online[ol.Vid]; ok {
		err = errors.New(fmt.Sprintf("vid %d exist", ol.Vid))
		return
	}
	ol.IPs = make(map[string]int64)
	ol.Total = 0
	srv.Online[ol.Vid] = &ol
	onlines = []*Online{}
	for _, online := range srv.Online {
		onlines = append(onlines, online)
	}
	return
}
