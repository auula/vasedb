package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

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
	s      *http.Server
	closed int32
	port   int
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
		port: opt.Port,
	}

	// 开启 HTTP/1.1 Keep-Alive 长连接
	hs.s.SetKeepAlivesEnabled(true)

	atomic.StoreInt32(&hs.closed, 0)

	return &hs
}

func (hs *HttpServer) Port() int {
	return hs.port
}

func (hs *HttpServer) Startup() error {

	if hs.port < 1024 || hs.port > 1<<16 {
		return errors.New("HTTP server port illegal")
	}

	if hs.closed != 0 {
		return errors.New("HTTP server has started")
	}

	atomic.StoreInt32(&hs.closed, 1)

	return hs.s.ListenAndServe()
}

func (hs *HttpServer) Shutdown() error {

	if hs.closed == 0 {
		return errors.New("HTTP server not started")
	}

	if err := hs.s.Shutdown(context.Background()); err != nil {
		return err
	}

	atomic.StoreInt32(&hs.closed, 1)

	return nil
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
