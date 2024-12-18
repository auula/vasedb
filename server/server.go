package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/vfs"
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

type Options struct {
	Port int
	Auth string
	Cert *tls.Config
	DB   *vfs.LogStructuredFS
}

// New 创建一个新的 HTTP 服务器
func New(opt *Options) (*HttpServer, error) {

	if opt.Port < minPort || opt.Port > maxPort {
		return nil, errors.New("HTTP server port illegal")
	}

	if opt.Auth != "" {
		authPassword = opt.Auth
	}

	hs := HttpServer{
		s: &http.Server{
			Handler:      root,
			Addr:         net.JoinHostPort(ipv4, strconv.Itoa(opt.Port)),
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
		},
		port:   opt.Port,
		closed: 0,
	}

	// 开启 HTTP Keep-Alive 长连接
	hs.s.SetKeepAlivesEnabled(true)

	atomic.StoreInt32(&hs.closed, 0)

	return &hs, nil
}

func SetupFS(store *vfs.LogStructuredFS) {
	storage = store
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
