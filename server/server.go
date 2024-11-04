package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/auula/vasedb/clog"
	vhttp "github.com/auula/vasedb/server/http"
)

// ipv4 return local IPv4 address
var ipv4 string = "127.0.0.1"

const (
	minPort = 1024
	maxPort = 1 << 16
	timeout = time.Second * 3
)

func init() {
	// 初始化的本地 IPv4 地址
	interfaces, err := net.Interfaces()
	if err != nil {
		clog.Errorf("get local IPv4 address failed: %s", err)
	}

	for _, face := range interfaces {
		adders, err := face.Addrs()
		if err != nil {
			clog.Errorf("get local IPv4 address failed: %s", err)
		}

		for _, addr := range adders {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ipv4 = ipNet.IP.String()
				}
			}
		}
	}

}

type HttpServer struct {
	s      *http.Server
	closed int32
	port   int
}

// New 创建一个新的 HTTP 服务器
func New(port int) (*HttpServer, error) {

	if port < minPort || port > maxPort {
		return nil, errors.New("HTTP server port illegal")
	}

	hs := HttpServer{
		s: &http.Server{
			// 因为 http 包是 server 子包，所以不需要提供行参传入，直接内嵌初始化
			Handler:      vhttp.Root,
			Addr:         net.JoinHostPort(ipv4, strconv.Itoa(port)),
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
		},
		port:   port,
		closed: 0,
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

	if hs.closed == 1 {
		return errors.New("HTTP server has started")
	}

	atomic.StoreInt32(&hs.closed, 1)

	return hs.s.ListenAndServe()
}

func (hs *HttpServer) Shutdown() error {

	if hs.closed == 0 {
		return errors.New("HTTP server not started")
	}

	err := hs.s.Shutdown(context.Background())
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	atomic.StoreInt32(&hs.closed, 0)

	return nil
}
