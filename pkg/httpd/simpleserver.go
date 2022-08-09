package httpd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/webhook"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

/*SimpleServer is the required config to create httpd server*/
type SimpleServer struct {
	Local    bool
	Port     int
	CertFile string
	KeyFile  string
	Patcher  webhook.SidecarInjectorPatcher
	Debug    bool
}

/*Start the simple http server supporting TLS*/
func (simpleServer *SimpleServer) Start() error {
	k8sClient, err := simpleServer.CreateClient()
	if err != nil {
		return err
	}

	simpleServer.Patcher.K8sClient = k8sClient
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", simpleServer.Port),
	}

	mux := http.NewServeMux()
	server.Handler = mux

	admissionHandler := &admission.Handler{
		Handler: &admission.PodAdmissionRequestHandler{
			PodHandler: &simpleServer.Patcher,
		},
	}
	mux.HandleFunc("/healthz", webhook.HealthCheckHandler)
	mux.HandleFunc("/mutate", admissionHandler.HandleAdmission)

	if simpleServer.Local {
		return server.ListenAndServe()
	}
	return server.ListenAndServeTLS(simpleServer.CertFile, simpleServer.KeyFile)
}

// CreateClient Create the server
func (simpleServer *SimpleServer) CreateClient() (*kubernetes.Clientset, error) {
	config, err := simpleServer.buildConfig()

	if err != nil {
		return nil, errors.Wrapf(err, "error setting up cluster config")
	}

	return kubernetes.NewForConfig(config)
}

func (simpleServer *SimpleServer) buildConfig() (*rest.Config, error) {
	if simpleServer.Local {
		log.Debug("Using local kubeconfig.")
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	log.Debug("Using in cluster kubeconfig.")
	return rest.InClusterConfig()
}
