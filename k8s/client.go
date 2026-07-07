// Package k8s builds Kubernetes clients that work both in-cluster and from a
// local kubeconfig.
package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset returns a Kubernetes clientset. It prefers in-cluster config and
// falls back to the kubeconfig at KUBECONFIG or ~/.kube/config.
func NewClientset() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = loadKubeconfig()
		if err != nil {
			return nil, fmt.Errorf("load kube config: %w", err)
		}
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build clientset: %w", err)
	}
	return cs, nil
}

// loadKubeconfig loads configuration from the local kubeconfig file.
func loadKubeconfig() (*rest.Config, error) {
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", path)
}
