package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
)

func NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	err := json.NewEncoder(w).Encode(ResponseData{
		Message: NotFoundMsg,
		Code:    http.StatusNotFound,
	})
	if err != nil {
		clog.Error(err)
		return
	}
}

func MethodNotAllowHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	err := json.NewEncoder(w).Encode(ResponseData{
		Message: MethodNotAllowedMsg,
		Code:    http.StatusMethodNotAllowed,
	})
	if err != nil {
		clog.Error(err)
		return
	}
}

func handleAction(w http.ResponseWriter, _ *http.Request) {
	resp := ResponseData{Result: []interface{}{
		"123 Main St, Any-town, CA, 12345",
	}, Status: "OK", Time: "0.34s"}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		clog.Error(err)
		return
	}
}

func handleRPC(w http.ResponseWriter, _ *http.Request) {

	resp := ResponseData{Result: []interface{}{
		Set{
			Key:  "value01",
			Data: []interface{}{68},
		},
		Set{
			Key:  "value02",
			Data: []interface{}{96, 46, 23, 12, 8},
		},
	}}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		clog.Error(err)
		return
	}
}

// 中间件函数，进行 BasicAuth 鉴权
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clog.Debug(fmt.Sprintf("%+v", conf.Settings))
		clog.Debug(fmt.Sprintf("HTTP request header authorization: %s", r.Header.Get("Auth")))

		var ipAddr string
		ipAddr = r.Header.Get("X-Forwarded-For")
		if ipAddr == "" {
			ipAddr = r.RemoteAddr
		}

		// 检查请求是否授权
		if r.Header.Get("Auth") != conf.Settings.Password {
			// 认证不成功，不继续处理请求
			clog.Warn(fmt.Sprintf("client %s connection to server failed", ipAddr))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ResponseData{
				Code:    http.StatusUnauthorized,
				Message: NotAuthorizationMsg,
			})
			if err != nil {
				clog.Error(err)
			}
			return
		}

		clog.Info(fmt.Sprintf("client %s connection to server success", ipAddr))
		next.ServeHTTP(w, r)
	})
}

func HandleStatus(w http.ResponseWriter, _ *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	err := json.NewEncoder(w).Encode(ResponseData{
		Code: http.StatusOK,
		Result: []interface{}{
			memStats,
		},
	})

	if err != nil {
		clog.Error(err)
	}
}
