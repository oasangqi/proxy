package http

import (
	"context"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func (s *Server) onlineVid(c *gin.Context) {
	var arg struct {
		Vid int64 `form:"vid" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	online, err := s.proxy.OnlineVid(c, arg.Vid)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res := map[string]interface{}{
		"vid":    arg.Vid,
		"online": *online,
	}
	result(c, res, OK)
}

func (s *Server) onlineAll(c *gin.Context) {
	online, err := s.proxy.OnlineAll(c)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res := map[string]interface{}{
		"online": online,
	}
	result(c, res, OK)
}

func (s *Server) addOnline(c *gin.Context) {
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	ols, err := s.proxy.AddOnline(context.TODO(), msg)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	res := map[string]interface{}{
		"online": ols,
	}
	result(c, res, OK)
}
