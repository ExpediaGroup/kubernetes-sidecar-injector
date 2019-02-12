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
	parameters := parse()

	mutator := &Mutator{
		props: parameters,
	}

	errs := make(chan error, 1)
	mutator.listen(errs)

	select {
	case err := <-errs:
		glog.Errorf("Filed to listen and serve mutator server: %v", err)
		os.Exit(1)
	case <-time.After(5 * time.Second):
		glog.Infof("Server listening in port %v", mutator.props.port)
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Shutting down initiated")
	mutator.shutdown(context.Background())
}

func parse() Parameters {
	var parameters Parameters

	flag.IntVar(&parameters.port, "port", 443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	flag.StringVar(&parameters.keyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	flag.Parse()

	return parameters
}
