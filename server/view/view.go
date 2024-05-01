package view

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

const (
	NotFoundMsg         = `The path you are requesting is not allowed!`
	MethodNotAllowedMsg = `Your access method is not allowed!`
	NotAuthorizationMsg = `Your request not authorization! add auth and password to your request header!`
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
