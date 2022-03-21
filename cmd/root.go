package cmd

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/httpd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	httpdConf         httpd.SimpleServer
	sideCarConfigFile string
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-sidecar-injector",
	Short: "Responsible for injecting sidecars into pod containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infof("SimpleServer starting to listen in port %v", httpdConf.Port)
		return httpdConf.Start(sideCarConfigFile)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVar(&httpdConf.Port, "port", 443, "server port.")
	rootCmd.Flags().StringVar(&httpdConf.CertFile, "certFile", "/etc/mutator/certs/cert.pem", "File containing tls certificate")
	rootCmd.Flags().StringVar(&httpdConf.KeyFile, "keyFile", "/etc/mutator/certs/key.pem", "File containing tls private key")
	rootCmd.Flags().BoolVar(&httpdConf.Local, "local", false, "Local run mode")
	rootCmd.Flags().StringVar(&sideCarConfigFile, "sideCar", "/etc/mutator/sidecar.yaml", "File containing sidecar template")
}
