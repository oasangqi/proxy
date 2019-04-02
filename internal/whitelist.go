package proxy

import "context"

func (srv *Server) ClearWhiteList(c context.Context) (err error) {
	srv.cfg.WebSocket.Debug = false
	srv.WhiteList = make(map[string]struct{})
	return
}

func (srv *Server) AddWhiteList(c context.Context, ips []string) (whitelist []string, err error) {
	srv.cfg.WebSocket.Debug = true
	for _, ip := range ips {
		srv.WhiteList[ip] = struct{}{}
	}
	return
}
