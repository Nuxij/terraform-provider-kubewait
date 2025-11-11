---
page_title: "kubewait Provider"
description: |-
  The kubewait provider allows you to wait for Kubernetes resources to meet specific conditions before proceeding with other Terraform operations.
---

# kubewait Provider

The kubewait provider allows you to wait for Kubernetes resources to meet specific conditions before proceeding with other Terraform operations. Perfect for ensuring cluster readiness, waiting for deployments, and controlling deployment order in complex Kubernetes environments.

## Features

- **Multiple Authentication Methods**: Auto-discovery, file paths, raw content, in-cluster
- **Per-Resource Configuration**: Each resource can authenticate to different clusters  
- **Zero Configuration**: Works out-of-the-box with standard Kubernetes setups
- Wait for Kubernetes resources to meet conditions (Ready, Available, etc.)
- Support for nodes, pods, deployments, services, jobs, and other resources
- Label and field selectors for precise resource targeting
- Configurable timeouts and check intervals
- Integration with Terraform dependency management (`depends_on`)

## Example Usage

```terraform
# Wait for all nodes to be ready
resource "kubewait_nodes" "cluster_ready" {
  for     = "condition=Ready"
  all     = true
  timeout = 600
}

# Wait for a specific deployment
resource "kubewait_deployments" "app_ready" {
  for       = "condition=Available"
  name      = "my-app"
  namespace = "production"
  timeout   = 300
}

# Generic wait for any Kubernetes resource
resource "kubewait_wait" "custom_resource" {
  for      = "condition=Ready"
  resource = "customresources"
  labels   = "app=my-operator"
  timeout  = 120
}
```

## Authentication

The provider supports multiple authentication methods:

```terraform
provider "kubewait" {
  # Option 1: Use kubeconfig file path
  kube_config_type = "file"
  kube_config      = "~/.kube/config"
  
  # Option 2: Use raw kubeconfig content
  kube_config_type = "raw"
  kube_config      = file("~/.kube/config")
  
  # Option 3: Auto-discovery (default)
  kube_config_type = "auto"  # Uses in-cluster or ~/.kube/config
  
  # Optional: Specify context
  context = "my-cluster-context"
  
  # Optional: Set default namespace
  namespace = "default"
}
```

## Schema

### Provider Configuration

- `kube_config_type` (String) Type of Kubernetes configuration. One of: "auto", "raw", "file", "provider". Defaults to "auto".
- `kube_config` (String, Sensitive) Kubernetes configuration content (when type is "raw") or file path (when type is "file").
- `context` (String) Kubernetes context to use from the kubeconfig.
- `namespace` (String) Default namespace for operations. Defaults to "default".