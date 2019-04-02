package http

import (
	"net/http"
	"proxy/internal"
	"proxy/internal/conf"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	proxy  *proxy.Server
	cfg    *conf.Config
}

func New(c *conf.Config, tp *proxy.Server) *Server {
	engine := gin.New()
	engine.Use(loggerHandler, recoverHandler)
	go func() {
		if err := engine.Run(c.Http.Addr); err != nil {
			panic(err)
		}
	}()
	s := &Server{
		engine: engine,
		proxy:  tp,
		cfg:    c,
	}
	s.initRouter()
	return s
}

func (s *Server) initRouter() {
	group := s.engine.Group("/proxy")
	group.GET("/online/vid", s.onlineVid)
	group.GET("/online/all", s.onlineAll)
	group.POST("/online/add", s.addOnline)
	group.GET("/whitelist/clear", s.clearWhitelist)
	group.GET("/whitelist/add", s.addWhiteList)
	s.engine.StaticFS("/proxy/log", http.Dir(s.cfg.Http.LogDir))
}

func (svr *Server) Close() {
}
