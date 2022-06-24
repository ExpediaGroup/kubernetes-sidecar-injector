package cmd

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/httpd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	httpdConf httpd.SimpleServer
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-sidecar-injector",
	Short: "Responsible for injecting sidecars into pod containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infof("SimpleServer starting to listen in port %v", httpdConf.Port)
		return httpdConf.Start()
	},
}

// Execute Kicks off the application
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
	rootCmd.Flags().StringVar(&(&httpdConf.Patcher).InjectPrefix, "injectPrefix", "sidecar-injector.expedia.com", "Injector Prefix")
	rootCmd.Flags().StringVar(&(&httpdConf.Patcher).InjectName, "injectName", "inject", "Injector Name")
	rootCmd.Flags().StringVar(&(&httpdConf.Patcher).SidecarDataKey, "sidecarDataKey", "sidecars.yaml", "ConfigMap Sidecar Data Key")
}
