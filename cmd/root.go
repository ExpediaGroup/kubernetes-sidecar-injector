package cmd

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/httpd"
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/routes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"os"
)

var (
	httpdConf         httpd.Conf
	sideCarConfigFile string
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-sidecar-injector",
	Short: "Responsible for injecting sidecars into pod containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		simpleServer := httpd.NewSimpleServer(httpdConf)

		if err := addRoutes(simpleServer); err != nil {
			return err
		}

		if err := startHTTPServerAndWait(simpleServer); err != nil {
			return err
		}

		glog.Infof("Shutting down initiated")
		simpleServer.Shutdown()
		return nil
	},
}

func addRoutes(simpleServer httpd.SimpleServer) error {
	mutator, err := routes.NewMutatorController(sideCarConfigFile)
	if err != nil {
		return err
	}

	simpleServer.AddRoute("/mutate", mutator.Mutate)
	return nil
}

func startHTTPServerAndWait(simpleServer httpd.SimpleServer) error {
	glog.Infof("SimpleServer starting to listen in port %v", simpleServer.Port())

	return simpleServer.Start()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		glog.Errorf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVar(&httpdConf.Port, "port", 443, "server port.")
	rootCmd.Flags().StringVar(&httpdConf.CertFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	rootCmd.Flags().StringVar(&httpdConf.KeyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	rootCmd.Flags().StringVar(&sideCarConfigFile, "sideCar", "/etc/mutator/sidecar.yaml", "File containing sidecar template")
}
