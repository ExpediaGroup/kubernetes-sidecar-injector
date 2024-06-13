package pkg

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	log "github.com/sirupsen/logrus"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

/*SimpleServer is the required config to create httpd server*/
type SimpleServer struct {
	Local           bool
	Port            int
	MetricsPort     int
	CertDir         string
	CertName        string
	KeyName         string
	SidecarInjector SidecarInjector
	Debug           bool
}

/*Start the simple http server supporting TLS*/
func (server *SimpleServer) Start() error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(server.Debug)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: fmt.Sprintf(":%d", server.MetricsPort),
		Port:               server.Port,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		return err
	}

	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = server.CertDir
	hookServer.CertName = server.CertName
	hookServer.KeyName = server.KeyName
	hookServer.Register("/mutate-v1-pod", &crwebhook.Admission{
		Handler: &SidecarInjector{},
	})
	hookServer.Register("/healthz", http.HandlerFunc(HealthCheckHandler))

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		return err
	}

	return nil
}
