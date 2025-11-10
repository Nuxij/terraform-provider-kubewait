# Basic Usage Examples

## Authentication Methods

The `kube_wait` provider supports multiple authentication methods:

### Auto-Discovery (Recommended for local development)

```hcl
provider "kube_wait" {}

# Uses in-cluster config (if running in pod) or ~/.kube/config
resource "kube_wait" "nodes_ready" {
  for      = "condition=Ready"
  resource = "nodes" 
  all      = true
  timeout  = 300
}
```

### Dynamic Kubeconfig from Data Sources

```hcl
provider "kube_wait" {}

# Get kubeconfig from cloud provider data source
resource "kube_wait" "cloud_nodes" {
  for         = "condition=Ready"
  resource    = "nodes"
  all         = true
  timeout     = 300
  kube_config = data.digitalocean_kubernetes_cluster.main.kube_config.0.raw_config
}
```

### Specific Kubeconfig File

```hcl
provider "kube_wait" {}

resource "kube_wait" "file_based" {
  for              = "condition=Ready"
  resource         = "nodes"
  all              = true
  timeout          = 300
  kube_config_path = "/path/to/cluster.kubeconfig"
  context          = "my-cluster-context"
}
```

### Provider-Level Defaults with Per-Resource Overrides

```hcl
provider "kube_wait" {
  kube_config_path = "~/.kube/config"
  context          = "default-cluster"
}

# Uses provider defaults
resource "kube_wait" "default_cluster" {
  for      = "condition=Ready"
  resource = "nodes"
  all      = true
}

# Overrides provider config for this resource
resource "kube_wait" "other_cluster" {
  for              = "condition=Ready"
  resource         = "nodes"
  all              = true
  kube_config_path = "./other-cluster.kubeconfig"
}
```

## Common Wait Patterns

### Wait for a Specific Pod

```hcl
resource "kube_wait" "nginx_pod" {
  for       = "condition=Ready"
  resource  = "pod"
  name      = "nginx-deployment-abc123"
  namespace = "default"
  timeout   = 120
  # Uses auto-discovery for auth
}
```

### Wait for Pods with Label Selector

```hcl
resource "kube_wait" "app_pods" {
  for         = "condition=Ready"
  resource    = "pod"
  labels      = "app=nginx,tier=frontend"
  namespace   = "production"
  all         = true
  timeout     = 180
  kube_config = data.eks_cluster.main.kubeconfig
}
```

## Wait for Deployment to be Available

```hcl
resource "kube_wait" "deployment_available" {
  for       = "condition=Available"
  resource  = "deployment"
  name      = "my-app"
  namespace = "default"
  timeout   = 300
}
```

## Wait for Pod Phase (Running)

```hcl
resource "kube_wait" "pod_running" {
  for       = "phase=Running"
  resource  = "pod"
  labels    = "job-name=data-migration"
  namespace = "jobs"
  timeout   = 600
}
```

## Using Field Selectors

```hcl
resource "kube_wait" "pods_on_node" {
  for            = "condition=Ready"
  resource       = "pod"
  field_selector = "spec.nodeName=worker-node-1"
  namespace      = "kube-system"
  timeout        = 120
}
```

## Provider Configuration Options

```hcl
provider "kube_wait" {
  # Use raw kubeconfig content
  kube_config = file("~/.kube/config")
  
  # Or specify path to kubeconfig file
  # kube_config_path = "~/.kube/config"
  
  # Use specific context
  # context = "my-cluster-context"
  
  # Set default namespace
  # namespace = "default"
}
```

## Complete Example with Dependencies

```hcl
terraform {
  required_providers {
    kube_wait = {
      source = "nuxij/kubewait"
      version = "~> 1.0"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kube_wait" {
  kube_config_path = "~/.kube/config"
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Create a deployment
resource "kubernetes_deployment" "nginx" {
  metadata {
    name = "nginx"
    labels = {
      app = "nginx"
    }
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "nginx"
      }
    }

    template {
      metadata {
        labels = {
          app = "nginx"
        }
      }

      spec {
        container {
          image = "nginx:1.21"
          name  = "nginx"

          port {
            container_port = 80
          }
        }
      }
    }
  }
}

# Wait for the deployment to be ready
resource "kube_wait" "nginx_ready" {
  for       = "condition=Available"
  resource  = "deployment"
  name      = "nginx"
  namespace = "default"
  timeout   = 300

  depends_on = [kubernetes_deployment.nginx]
}

# Create a service only after deployment is ready
resource "kubernetes_service" "nginx" {
  metadata {
    name = "nginx"
  }

  spec {
    selector = {
      app = "nginx"
    }

    port {
      port        = 80
      target_port = 80
    }

    type = "LoadBalancer"
  }

  depends_on = [kube_wait.nginx_ready]
}
```