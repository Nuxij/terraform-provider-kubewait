# Example Usage: DigitalOcean Kubernetes Cluster with Helm

This example shows multiple ways to use the `kube_wait` provider with different authentication methods.

## Option 1: Dynamic Kubeconfig from Data Source

```hcl
terraform {
  required_providers {
    kube_wait = {
      source = "nuxij/kubewait"
      version = "~> 1.0"
    }
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Configure the DigitalOcean Provider
provider "digitalocean" {
  token = var.do_token
}

# Reference an existing DigitalOcean Kubernetes cluster
data "digitalocean_kubernetes_cluster" "existing" {
  name = var.cluster_name
}

# No provider config needed - each resource configures itself
provider "kube_wait" {}

provider "helm" {
  kubernetes {
    host     = data.digitalocean_kubernetes_cluster.existing.endpoint
    token    = data.digitalocean_kubernetes_cluster.existing.kube_config.0.token
    cluster_ca_certificate = base64decode(
      data.digitalocean_kubernetes_cluster.existing.kube_config.0.cluster_ca_certificate
    )
  }
}

# Wait for all nodes using dynamic kubeconfig from data source
resource "kube_wait" "all-nodes" {
  for         = "condition=Ready"
  resource    = "nodes"
  all         = true
  timeout     = 600
  kube_config = data.digitalocean_kubernetes_cluster.existing.kube_config.0.raw_config
}

# Wait for DO agent using same dynamic kubeconfig
resource "kube_wait" "do-agent-pod" {
  for         = "condition=Ready"
  resource    = "pod"
  labels      = "app=do-node-agent"
  namespace   = "kube-system"
  timeout     = 600
  kube_config = data.digitalocean_kubernetes_cluster.existing.kube_config.0.raw_config
}
```

## Option 2: Auto-Discovery (In-Cluster or ~/.kube/config)

```hcl
# When running from within a Kubernetes pod or with ~/.kube/config available
provider "kube_wait" {}

# Auto-discovers kubeconfig - tries in-cluster first, then ~/.kube/config
resource "kube_wait" "cluster_ready" {
  for      = "condition=Ready"
  resource = "nodes"
  all      = true
  timeout  = 600
  # No kube_config specified - uses auto-discovery
}

resource "kube_wait" "system_pods" {
  for       = "condition=Ready"
  resource  = "pod"
  labels    = "k8s-app=kube-dns"
  namespace = "kube-system"
  timeout   = 300
  # Also uses auto-discovery
}
```

## Option 3: Specific Kubeconfig File Path

```hcl
provider "kube_wait" {}

# Use a specific kubeconfig file
resource "kube_wait" "production_nodes" {
  for              = "condition=Ready"
  resource         = "nodes"
  all              = true
  timeout          = 600
  kube_config_path = "/path/to/production/kubeconfig"
  context          = "production-cluster"
}

# Different cluster with different config
resource "kube_wait" "staging_nodes" {
  for              = "condition=Ready" 
  resource         = "nodes"
  all              = true
  timeout          = 300
  kube_config_path = "/path/to/staging/kubeconfig"
  context          = "staging-cluster"
}
```

## Option 4: Mixed Authentication Methods

```hcl
provider "kube_wait" {
  # Optional: Set defaults that resources can inherit
  kube_config_path = "~/.kube/config"
  context          = "default-cluster"
}

# This resource inherits provider defaults
resource "kube_wait" "default_cluster" {
  for      = "condition=Ready"
  resource = "nodes"
  all      = true
}

# This resource overrides with dynamic config
resource "kube_wait" "remote_cluster" {
  for         = "condition=Ready"
  resource    = "nodes" 
  all         = true
  kube_config = data.digitalocean_kubernetes_cluster.remote.kube_config.0.raw_config
}

# This resource uses a specific file
resource "kube_wait" "test_cluster" {
  for              = "condition=Ready"
  resource         = "nodes"
  all              = true
  kube_config_path = "./test-cluster.kubeconfig"
}
```

# Install Prometheus only after cluster is fully ready
resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  namespace  = "monitoring"
  create_namespace = true

  values = [
    yamlencode({
      grafana = {
        enabled = true
        adminPassword = "admin"
      }
    })
  ]

  depends_on = [
    kube_wait.all-nodes,
    kube_wait.do-agent-pod,
  ]
}

# Variables
variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "cluster_name" {
  description = "Name of the existing DigitalOcean Kubernetes cluster"
  type        = string
  default     = "my-k8s-cluster"
}
```

## Alternative Approach: Two-Step Deployment

If you need to create a new cluster, use a two-step approach:

### Step 1: Create the cluster (cluster.tf)

```hcl
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_kubernetes_cluster" "example" {
  name    = "example-cluster"
  region  = "nyc3"
  version = "1.28.2-do.0"

  node_pool {
    name       = "default"
    size       = "s-2vcpu-2gb"
    node_count = 3
  }
}

# Output the cluster name for the next step
output "cluster_name" {
  value = digitalocean_kubernetes_cluster.example.name
}
```

### Step 2: Deploy applications with wait logic (apps.tf)

```hcl
terraform {
  required_providers {
    kube_wait = {
      source = "nuxij/kubewait"
      version = "~> 1.0"
    }
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Get the cluster created in step 1
data "digitalocean_kubernetes_cluster" "existing" {
  name = var.cluster_name  # From step 1 output
}

# Configure providers
provider "kube_wait" {
  kube_config = data.digitalocean_kubernetes_cluster.existing.kube_config.0.raw_config
}

provider "helm" {
  kubernetes {
    host     = data.digitalocean_kubernetes_cluster.existing.endpoint
    token    = data.digitalocean_kubernetes_cluster.existing.kube_config.0.token
    cluster_ca_certificate = base64decode(
      data.digitalocean_kubernetes_cluster.existing.kube_config.0.cluster_ca_certificate
    )
  }
}

# Wait for cluster readiness, then deploy apps
resource "kube_wait" "cluster_ready" {
  for       = "condition=Ready"
  resource  = "nodes"
  all       = true
  timeout   = 600
}

resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  namespace  = "monitoring"
  create_namespace = true

  depends_on = [kube_wait.cluster_ready]
}
```

## Authentication Priority Order

The provider uses this order to find Kubernetes configuration:

1. **Resource-level `kube_config_type`** (if explicitly set) - highest priority
2. **Provider-level configuration** (if `kube_config_type` is omitted and provider config exists)
3. **Auto-discovery fallback** (in-cluster config, then `~/.kube/config`)

### Detailed Precedence:

- If `kube_config_type = "raw"` → Use `kube_config` as raw content
- If `kube_config_type = "file"` → Use `kube_config` as file path  
- If `kube_config_type = "auto"` → Use auto-discovery (ignore provider)
- If `kube_config_type = "provider"` → Use provider config
- If `kube_config_type` is **omitted**:
  - **First**: Check if provider has `kube_config` or `kube_config_path` → Use provider
  - **Then**: Fallback to auto-discovery

## Key Benefits

1. **Flexibility**: Each resource can authenticate to different clusters
2. **No Provider Dependencies**: Resources handle their own authentication  
3. **Zero Configuration**: Works out of the box with standard Kubernetes setups
4. **Dynamic Configuration**: Can use data sources for kubeconfig content
5. **Multi-Cluster Support**: Different resources can target different clusters in the same configuration

## Variables

```hcl
variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "cluster_name" {
  description = "Name of the existing DigitalOcean Kubernetes cluster"
  type        = string
  default     = "my-k8s-cluster"
}
```