package cmd

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	server pkg.SimpleServer
	debug  bool
)

var rootCmd = &cobra.Command{
	Use:   "kubernetes-sidecar-injector",
	Short: "Responsible for injecting sidecars into pod containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
		log.Infof("SimpleServer starting to listen in port %v", server.Port)
		return server.Start()
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
	rootCmd.Flags().IntVar(&server.Port, "port", 443, "server port.")
	rootCmd.Flags().IntVar(&server.MetricsPort, "metricsPort", 9090, "metrics server port.")
	rootCmd.Flags().StringVar(&server.CertDir, "certDir", "/etc/mutator/certs/", "The directory that contains the server key and certificate")
	rootCmd.Flags().StringVar(&server.CertName, "certName", "tls.crt", "The server certificate name")
	rootCmd.Flags().StringVar(&server.KeyName, "keyName", "tls.key", "The server key name")
	rootCmd.Flags().BoolVar(&server.Local, "local", false, "Local run mode")
	rootCmd.Flags().StringVar(&(&server.SidecarInjector).InjectPrefix, "injectPrefix", "sidecar-injector.expedia.com", "Injector Prefix")
	rootCmd.Flags().StringVar(&(&server.SidecarInjector).InjectName, "injectName", "inject", "Injector Name")
	rootCmd.Flags().StringVar(&(&server.SidecarInjector).SidecarDataKey, "sidecarDataKey", "sidecars.yaml", "ConfigMap Sidecar Data Key")
	rootCmd.Flags().BoolVar(&(&server.SidecarInjector).AllowAnnotationOverrides, "allowAnnotationOverrides", false, "Allow Annotation Overrides")
	rootCmd.Flags().BoolVar(&(&server.SidecarInjector).AllowLabelOverrides, "allowLabelOverrides", false, "Allow Label Overrides")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "enable debug logs")
}
