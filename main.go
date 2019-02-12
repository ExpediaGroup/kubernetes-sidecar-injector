package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	parameters := parseArguments()

	if mutator, ok := startHttpsServer(parameters); ok {
		wait()
		mutator.shutdown(context.Background())
	} else {
		os.Exit(1)
	}
}

func parseArguments() Parameters {
	var parameters Parameters

	flag.IntVar(&parameters.port, "port", 443, "server port.")
	flag.StringVar(&parameters.certFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	flag.StringVar(&parameters.keyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	flag.Parse()

	return parameters
}

func startHttpsServer(parameters Parameters) (*Mutator, bool) {
	mutator := &Mutator{
		props: parameters,
	}

	errs := make(chan error, 1)
	mutator.listen(errs)

	select {
	case err := <-errs:
		glog.Errorf("Filed to listen and serve mutator server: %v", err)
		return nil, false
	case <-time.After(5 * time.Second):
		glog.Infof("Server listening in port %v", mutator.props.port)
	}

	return mutator, true
}

func wait() {
	// subscribe to process shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Shutting down initiated")
}
