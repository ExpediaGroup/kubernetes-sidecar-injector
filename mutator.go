package main

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	//corev1 "k8s.io/api/core/v1"
	"net/http"
)

//type Sidecar struct {
//	Containers  []corev1.Container  `yaml:"containers"`
//	Volumes     []corev1.Volume     `yaml:"volumes"`
//}

type Parameters struct {
	port     int
	certFile string
	keyFile  string
}

type Mutator struct {
	props Parameters
}

var server *http.Server

func (m *Mutator) listen(errs chan error) {
	server = &http.Server{
		Addr: fmt.Sprintf(":%v", m.props.port),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", m.serve)
	server.Handler = mux

	go func() {
		if err := server.ListenAndServeTLS(m.props.certFile, m.props.keyFile); err != nil {
			errs <- err
		}
		close(errs)
	}()
}

func (m *Mutator) serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello World!")); err != nil {
		glog.Errorf("Failed writing response: %v", err)
		http.Error(w, fmt.Sprintf("Failed writing response: %v", err), http.StatusInternalServerError)
	}
}

func (m *Mutator) shutdown(background context.Context) {
	_ = server.Shutdown(background)
}
