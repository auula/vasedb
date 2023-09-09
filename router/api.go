package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

var (
	Root   *mux.Router
	sys    = []string{"GET", "POST"}
	curd   = []string{"POST", "DELETE", "PUT"}
	lambda = []string{"POST", "GET"}
)

func init() {
	Root = mux.NewRouter()
	Root.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	Root.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowHandler)
	Root.HandleFunc("/ql", handleRPC).Methods(curd...)
	Root.HandleFunc("/status", HandleStatus).Methods(sys...)
	Root.HandleFunc("/lambda", handleAction).Methods(lambda...)
	Root.Use(authMiddleware)
}

type Set struct {
	Key  string
	Data []interface{}
}

type ResponseData struct {
	Code    int           `json:"code"`
	Status  string        `json:"status"`
	Time    string        `json:"time"`
	Result  []interface{} `json:"result"`
	Message string        `json:"messages"`
}
