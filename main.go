package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/mchandramouli/haystack-kube-sidecar-injector/httpd"
	"github.com/mchandramouli/haystack-kube-sidecar-injector/routes"
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

	if err = startHttpsServer(simpleServer); err != nil {
		return
	}

	glog.Infof("SimpleServer listening in port %v", simpleServer.Port())
	wait(func() {
		glog.Infof("Shutting down initiated")
		simpleServer.Shutdown()
	})
	return
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

func startHttpsServer(simpleServer httpd.SimpleServer) error {
	errC := make(chan error, 1)
	defer close(errC)
	simpleServer.Start(errC)
	err := <-errC
	return err
}

func wait(callback func()) {
	// subscribe to process shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	callback()
}
