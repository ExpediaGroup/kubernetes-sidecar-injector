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

func NewSimpleServer(conf Conf) SimpleServer {
	simpleServer := simpleServerImpl{
		conf: conf,
	}
	simpleServer.server = &http.Server{
		Addr: fmt.Sprintf(":%v", simpleServer.conf.Port),
	}
	simpleServer.mux = http.NewServeMux()
	return simpleServer
}

type simpleServerImpl struct {
	conf Conf
	server *http.Server
	mux *http.ServeMux
}

func (simpleServer simpleServerImpl) Port() int {
	return simpleServer.conf.Port
}

func (simpleServer simpleServerImpl) AddRoute(pattern string, route Route) {
	simpleServer.mux.HandleFunc(pattern, route)
}

func (simpleServer simpleServerImpl) Start(errs chan error) {
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

func (simpleServer simpleServerImpl) Shutdown() {
	_ = simpleServer.server.Shutdown(context.Background())
}
