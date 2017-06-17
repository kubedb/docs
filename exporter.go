package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/appscode/log"
	"github.com/appscode/pat"
	"github.com/k8sdb/apimachinery/pkg/analytics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func NewCmdExport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Prometheus metrics for HAProxy",
		Run: func(cmd *cobra.Command, args []string) {
			export()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			analytics.SendEvent("exporter", "stopped", Version)
		},
	}

	// operator flags
	cmd.Flags().StringVar(&masterURL, "master", masterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&address, "address", address, "Address to listen on for web interface and telemetry.")

	// Analytics flags
	cmd.Flags().BoolVar(&enableAnalytics, "analytics", enableAnalytics, "Send analytical event to Google Analytics")

	return cmd
}

func export() {
	fmt.Println("Starting exporter...")

	if enableAnalytics {
		analytics.Enable()
	}
	analytics.SendEvent("exporter", "started", Version)

	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
	pattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", PathParamNamespace, PathParamType, PathParamName)
	log.Infoln("URL pattern:", pattern)
	m.Get(pattern, http.HandlerFunc(ExportMetrics))
	m.Del(pattern, http.HandlerFunc(DeleteRegistry))
	http.Handle("/", m)

	log.Infof("Starting Server: %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
