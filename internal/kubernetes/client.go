package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes clientset with additional functionality
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// ClientConfig holds configuration for creating a Kubernetes client
type ClientConfig struct {
	KubeConfig     string // Raw kubeconfig content
	KubeConfigPath string // Path to kubeconfig file
	Context        string // Kubernetes context to use
}

// expandPath expands ~ to home directory in file paths
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home := homedir.HomeDir(); home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// NewClient creates a new Kubernetes client based on the provided configuration
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	var kubeConfig *rest.Config
	var err error

	// Determine how to load the kubeconfig
	if config.KubeConfig != "" {
		// Use raw kubeconfig content
		kubeConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(config.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to create config from raw kubeconfig: %w", err)
		}
	} else {
		// Use kubeconfig file
		kubeconfigPath := expandPath(config.KubeConfigPath)
		if kubeconfigPath == "" {
			// Try in-cluster config first
			kubeConfig, err = rest.InClusterConfig()
			if err != nil {
				// Fall back to default kubeconfig path
				if home := homedir.HomeDir(); home != "" {
					kubeconfigPath = filepath.Join(home, ".kube", "config")
				}
			}
		}

		if kubeConfig == nil {
			// Load from file
			if kubeconfigPath == "" {
				return nil, fmt.Errorf("no kubeconfig path provided and unable to determine default path")
			}

			// Check if file exists
			if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfigPath)
			}

			// Build config from file
			configLoadingRules := &clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfigPath,
			}

			configOverrides := &clientcmd.ConfigOverrides{}
			if config.Context != "" {
				configOverrides.CurrentContext = config.Context
			}

			kubeConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				configLoadingRules,
				configOverrides,
			).ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to create config from kubeconfig file %s: %w", kubeconfigPath, err)
			}
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    kubeConfig,
	}, nil
}
