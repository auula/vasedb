package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/server/api"
)

var (
	// IPv4 return local IPv4 address
	// - get it once globally,
	// - don't try to get it dynamically on the fly.
	IPv4 = LocalIPv4Address()
)

type HttpServer struct {
	s        *http.Server
	shutdown chan struct{}
	closed   int32
	port     int
}

// New 创建一个新的 HTTP 服务器
func New(opt *conf.ServerConfig) *HttpServer {

	hs := HttpServer{
		s: &http.Server{
			Handler:      api.Root,
			Addr:         net.JoinHostPort(IPv4, strconv.Itoa(opt.Port)),
			WriteTimeout: 3 * time.Second,
			ReadTimeout:  3 * time.Second,
		},
		port:     opt.Port,
		shutdown: make(chan struct{}),
	}

	atomic.StoreInt32(&hs.closed, 0)

	return &hs
}

func (hs *HttpServer) Startup() {

	if hs.port < 1024 || hs.port > 1<<16 {
		clog.Failed("HTTP server port illegal")
	}

	if hs.closed != 0 {
		clog.Failed("HTTP server has started")
	}

	go func() {
		clog.Info("Receiving client connections")
		if err := hs.s.ListenAndServe(); err != nil {
			clog.Failed(err)
		}
	}()

	// 防止 HTTP 端口占用，延迟输出启动信息
	time.Sleep(500 * time.Millisecond)
	clog.Info(fmt.Sprintf("HTTP server started %s:%d 🚀", IPv4, hs.port))
	atomic.StoreInt32(&hs.closed, 1)

	<-hs.shutdown
}

func (hs *HttpServer) Shutdown() {

	if hs.closed == 0 {
		clog.Failed("HTTP server not started")
	}

	if err := hs.s.Shutdown(context.Background()); err != nil {
		clog.Failed(err)
	}

	atomic.StoreInt32(&hs.closed, 1)
	hs.shutdown <- struct{}{}
	close(hs.shutdown)

	clog.Info("Shutting down http server")
}

// LocalIPv4Address 返回本地 IPv4 地址
func LocalIPv4Address() string {
	var ip = "127.0.0.1"

	interfaces, err := net.Interfaces()
	if err != nil {
		return ip
	}

	for _, face := range interfaces {
		adders, err := face.Addrs()
		if err != nil {
			return ip
		}

		for _, addr := range adders {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}

	return ip
}
