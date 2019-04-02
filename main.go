package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"proxy/internal"
	"proxy/internal/conf"
	"proxy/internal/http"
	"syscall"

	"github.com/golang/glog"
)

func main() {
	var (
		srv     *proxy.Server
		httpSvr *http.Server
		err     error
	)
	flag.Parse()
	if err = conf.Init(); err != nil {
		panic(err)
	}
	if err = writePidFile(conf.Conf.PidFile); err != nil {
		panic(err)
	}
	if srv, err = proxy.New(conf.Conf); err != nil {
		panic(err)
	}
	if conf.Conf.Http.Open {
		httpSvr = http.New(conf.Conf, srv)
	}
	srv.Run()
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		glog.Infof("proxy get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if conf.Conf.Http.Open {
				httpSvr.Close()
			}
			srv.Close()
			glog.Infof("proxy [version: 1.0.0] exit")
			glog.Flush()
			return
		case syscall.SIGHUP:
			glog.Flush()
		default:
			return
		}
	}
}

func writePidFile(pidFile string) (err error) {
	pf, err := os.OpenFile(pidFile, os.O_RDWR, 0)
	defer pf.Close()
	if os.IsNotExist(err) {
		if pf, err = os.Create(pidFile); err != nil {
			glog.Errorf("create pid file:%s failed", pidFile)
			return
		}
	} else if err != nil {
		glog.Errorf("open pid file:%s failed", pidFile)
		return
	}
	pid := os.Getpid()
	_, err = pf.Write([]byte(fmt.Sprintf("%d", pid)))
	if err != nil {
		glog.Errorf("write to pid file:%s failed", pidFile)
		return
	}
	return
}
