package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

func (s *Server) clearWhitelist(c *gin.Context) {
	err := s.proxy.ClearWhiteList(context.TODO())
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	result(c, nil, OK)
}

func (s *Server) addWhiteList(c *gin.Context) {
	var arg struct {
		Ips []string `form:"ips" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	ips, err := s.proxy.AddWhiteList(context.TODO(), arg.Ips)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res := map[string]interface{}{
		"whitelist": ips,
	}
	result(c, res, OK)
}
