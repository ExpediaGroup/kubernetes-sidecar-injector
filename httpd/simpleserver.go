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
	return &simpleServerImpl{
		conf: conf,
		mux:  http.NewServeMux(),
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", conf.Port),
		},
	}
}

type simpleServerImpl struct {
	conf   Conf
	server *http.Server
	mux    *http.ServeMux
}

func (s *simpleServerImpl) Port() int {
	return s.conf.Port
}

func (s *simpleServerImpl) AddRoute(pattern string, route Route) {
	s.mux.HandleFunc(pattern, route)
}

func (s *simpleServerImpl) Start(errs chan error) {
	s.server.Handler = s.mux
	go func() {
		if err := s.server.ListenAndServeTLS(
			s.conf.CertFile,
			s.conf.KeyFile); err != nil {
			errs <- err
		}
	}()
}

func (s *simpleServerImpl) Shutdown() {
	s.server.Shutdown(context.Background())
}
