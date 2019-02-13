package httpd

import (
	"context"
	"fmt"
	"net/http"
)

type Conf struct {
	Port     int
	CertFile string
	KeyFile  string
}

type Route func(http.ResponseWriter, *http.Request)

type SimpleServer interface {
	Port() int
	AddRoute(string, Route)
	Start(chan error)
	Shutdown()
}

func NewServer(conf Conf) SimpleServer {
	simpleServer := lightServer{
		conf: conf,
	}
	simpleServer.server = &http.Server{
		Addr: fmt.Sprintf(":%v", simpleServer.conf.Port),
	}
	simpleServer.mux = http.NewServeMux()
	return simpleServer
}

type lightServer struct {
	conf Conf
	server *http.Server
	mux *http.ServeMux
}

func (simpleServer lightServer) Port() int {
	return simpleServer.conf.Port
}

func (simpleServer lightServer) AddRoute(pattern string, route Route) {
	simpleServer.mux.HandleFunc(pattern, route)
}

func (simpleServer lightServer) Start(errs chan error) {
	simpleServer.server.Handler = simpleServer.mux
	go func() {
		if err := simpleServer.server.ListenAndServeTLS(
			simpleServer.conf.CertFile,
			simpleServer.conf.KeyFile); err != nil {
			errs <- err
		}
		close(errs)
	}()
}

func (simpleServer lightServer) Shutdown() {
	_ = simpleServer.server.Shutdown(context.Background())
}
