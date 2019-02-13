package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/mchandramouli/haystack-kube-sidecar-injector/httpd"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	httpdConf := readHttpdConf()

	simpleServer := httpd.NewServer(httpdConf)

	simpleServer.AddRoute("/mutate", serve)

	if startHttpsServer(simpleServer) {
		wait(func() {
			glog.Infof("Shutting down initiated")
			simpleServer.Shutdown()
		})
	} else {
		os.Exit(1)
	}
}

func readHttpdConf() httpd.Conf {
	var httpdConf httpd.Conf

	flag.IntVar(&httpdConf.Port, "port", 443, "server port.")
	flag.StringVar(&httpdConf.CertFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	flag.StringVar(&httpdConf.KeyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	flag.Parse()

	return httpdConf
}

func startHttpsServer(simpleServer httpd.SimpleServer) bool {
	errs := make(chan error, 1)
	simpleServer.Start(errs)

	select {
	case err := <-errs:
		glog.Errorf("Filed to listen and serve : %v", err)
		return false
	case <-time.After(5 * time.Second):
		glog.Infof("SimpleServer listening in port %v", simpleServer.Port())
	}

	return true
}

func wait(callback func()) {
	// subscribe to process shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	callback()
}

func serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello World!")); err != nil {
		glog.Errorf("Failed writing response: %v", err)
		http.Error(w, fmt.Sprintf("Failed writing response: %v", err), http.StatusInternalServerError)
	}
}
