package httpd

import (
	"fmt"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/routes"
	"net/http"
)

/*SimpleServer is the required config to create httpd server*/
type SimpleServer struct {
	Port     int
	CertFile string
	KeyFile  string
}

/*Start the simple http server supporting TLS*/
func (conf *SimpleServer) Start(sideCarConfigFile string) error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", conf.Port),
	}

	mux := http.NewServeMux()
	server.Handler = mux

	if mutator, err := routes.NewMutatorController(sideCarConfigFile); err != nil {
		return err
	} else {
		mux.HandleFunc("/mutate", mutator.Mutate)
	}

	return server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
}
