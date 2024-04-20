package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/auula/vasedb/server/api"
)

var (
	// ipv4 return local IPv4 address
	// - 只读模式第一次获取之后就不需要改变.
	ipv4 = func() string {
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
	}()
)

const timeout = time.Second * 3

type HttpServer struct {
	s      *http.Server
	closed int32
	port   int
}

// New 创建一个新的 HTTP 服务器
func New(port int) (*HttpServer, error) {

	if port < 1024 || port > 1<<16 {
		return nil, errors.New("HTTP server port illegal")
	}

	hs := HttpServer{
		s: &http.Server{
			Handler:      api.Root,
			Addr:         net.JoinHostPort(ipv4, strconv.Itoa(port)),
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
		},
		port: port,
	}

	// 开启 HTTP Keep-Alive 长连接
	hs.s.SetKeepAlivesEnabled(true)

	atomic.StoreInt32(&hs.closed, 0)

	return &hs, nil
}

func (hs *HttpServer) Port() int {
	return hs.port
}

// IPv4 return local IPv4 address
func (hs *HttpServer) IPv4() string {
	return ipv4
}

// Startup blocking goroutine
func (hs *HttpServer) Startup() error {

	if !(hs.closed == 0) {
		return errors.New("HTTP server has started")
	}

	atomic.StoreInt32(&hs.closed, 1)

	return hs.s.ListenAndServe()
}

func (hs *HttpServer) Shutdown() error {

	if !(hs.closed == 1) {
		return errors.New("HTTP server not started")
	}

	err := hs.s.Shutdown(context.Background())
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	atomic.StoreInt32(&hs.closed, 0)

	return nil
}
