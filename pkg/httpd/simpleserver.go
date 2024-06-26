package httpd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/webhook"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

/*SimpleServer is the required config to create httpd server*/
type SimpleServer struct {
	Local       bool
	Port        int
	MetricsPort int
	CertFile    string
	KeyFile     string
	Patcher     webhook.SidecarInjectorPatcher
	Debug       bool
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

	metricsHandler := promhttp.Handler()
	if simpleServer.MetricsPort != simpleServer.Port {
		go simpleServer.startMetricsServer(metricsHandler)
	} else {
		mux.Handle("/metrics", metricsHandler)
	}

	if simpleServer.Local {
		return server.ListenAndServe()
	}
	return server.ListenAndServeTLS(simpleServer.CertFile, simpleServer.KeyFile)
}

func (simpleServer *SimpleServer) startMetricsServer(metricsHandler http.Handler) {
	log.Printf("Starting metrics server on port %d\n", simpleServer.MetricsPort)
	metricsRouter := http.NewServeMux()
	metricsRouter.Handle("/metrics", metricsHandler)

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", simpleServer.MetricsPort),
		Handler: metricsRouter,
	}

	if err := metricsServer.ListenAndServe(); err != nil {
		log.Fatal("Failed to start metrics server:", err)
	}
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
