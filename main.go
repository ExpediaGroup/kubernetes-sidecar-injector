package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ExpediaDotCom/kubernetes-sidecar-injector/httpd"
	"github.com/ExpediaDotCom/kubernetes-sidecar-injector/routes"
	"github.com/golang/glog"
)

type config struct {
	httpdConf         httpd.Conf
	sideCarConfigFile string
}

func main() {
	conf := readConfig()
	simpleServer := httpd.NewSimpleServer(conf.httpdConf)

	var err error
	defer func() {
		if err != nil {
			glog.Errorf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	if err = addRoutes(simpleServer, conf); err != nil {
		return
	}

	if err = startHTTPServerAndWait(simpleServer); err != nil {
		return
	}

	glog.Infof("Shutting down initiated")
	simpleServer.Shutdown()
}

func addRoutes(simpleServer httpd.SimpleServer, conf config) error {
	mutator, err := routes.NewMutatorController(conf.sideCarConfigFile)
	if err != nil {
		return err
	}

	simpleServer.AddRoute("/mutate", mutator.Mutate)
	return nil
}

func readConfig() config {
	var conf config

	flag.IntVar(&conf.httpdConf.Port, "port", 443, "server port.")
	flag.StringVar(&conf.httpdConf.CertFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	flag.StringVar(&conf.httpdConf.KeyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	flag.StringVar(&conf.sideCarConfigFile, "sideCar", "/etc/mutator/sidecar.yaml", "File containing sidecar template")
	flag.Parse()

	return conf
}

func startHTTPServerAndWait(simpleServer httpd.SimpleServer) error {
	errC := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		close(errC)
		close(signalChan)
	}()

	glog.Infof("SimpleServer starting to listen in port %v", simpleServer.Port())

	simpleServer.Start(errC)

	// block until an error or signal from os to
	// terminate the process
	var retErr error
	select {
	case err := <-errC:
		retErr = err
	case <-signalChan:
	}

	return retErr
}
