package httpd

import (
	"fmt"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/webhook"
	"net/http"
)

/*SimpleServer is the required config to create httpd server*/
type SimpleServer struct {
	Local    bool
	Port     int
	CertFile string
	KeyFile  string
	Patcher  webhook.SidecarInjectorPatcher
}

/*Start the simple http server supporting TLS*/
func (conf *SimpleServer) Start() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", conf.Port),
	}

	mux := http.NewServeMux()
	server.Handler = mux

	admissionHandler := &admission.Handler{
		Handler: &admission.PodAdmissionRequestHandler{
			PodHandler: &conf.Patcher,
		},
	}
	mux.HandleFunc("/mutate", admissionHandler.HandleAdmission)

	if conf.Local {
		return server.ListenAndServe()
	} else {
		return server.ListenAndServeTLS(conf.CertFile, conf.KeyFile)
	}
}
