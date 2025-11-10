# Terraform Provider: kube_wait

A Terraform provider that allows you to wait for Kubernetes resources to meet specific conditions before proceeding with other Terraform operations. Perfect for ensuring cluster readiness, waiting for deployments, and controlling deployment order in complex Kubernetes environments.

## Features

- **Multiple Authentication Methods**: Auto-discovery, file paths, raw content, in-cluster
- **Per-Resource Configuration**: Each resource can authenticate to different clusters  
- **Zero Configuration**: Works out-of-the-box with standard Kubernetes setups
- Wait for Kubernetes resources to meet conditions (Ready, Available, etc.)
- Support for nodes, pods, deployments, services, jobs, and other resources
- Label and field selectors for precise resource targeting
- Configurable timeouts and check intervals
- Integration with Terraform dependency management (`depends_on`)
- Context-aware operations

## Quick Start

### Basic Resource Waiting

Wait for all nodes to be ready:
```hcl
resource "kubewait_nodes" "cluster_ready" {
  for     = "condition=Ready"
  all     = true
  timeout = 600
}
```

Wait for a specific deployment:
```hcl
resource "kubewait_deployments" "app_ready" {
  for       = "condition=Available"
  name      = "my-app"
  namespace = "production"
  timeout   = 300
}
```

### Generic Wait Resource

For any Kubernetes resource type:
```hcl
resource "kubewait_wait" "custom_resource" {
  for      = "condition=Ready"
  resource = "customresources"
  labels   = "app=my-operator"
  timeout  = 120
}
```

## Available Resources

### Specific Resource Types
- `kubewait_nodes` - Wait for cluster nodes
- `kubewait_pods` - Wait for pods
- `kubewait_deployments` - Wait for deployments  
- `kubewait_services` - Wait for services
- `kubewait_jobs` - Wait for jobs
- `kubewait_cronjobs` - Wait for cronjobs
- `kubewait_daemonsets` - Wait for daemonsets
- `kubewait_statefulsets` - Wait for statefulsets
- `kubewait_ingress` - Wait for ingress resources

### Generic Resource
- `kubewait_wait` - Wait for any Kubernetes resource type

## Common Use Cases

### DigitalOcean Kubernetes Clusters
Wait for DigitalOcean's system components before deploying applications:

```hcl
resource "kubewait_pods" "do_agent" {
  for         = "condition=Ready"
  labels      = "app=do-node-agent"
  namespace   = "kube-system"
  timeout     = 600
  kube_config = data.digitalocean_kubernetes_cluster.main.kube_config.0.raw_config
}

resource "helm_release" "prometheus" {
  # ... helm configuration ...
  depends_on = [kubewait_pods.do_agent]
}
```

### Deployment Orchestration
Ensure applications start in the correct order:

```hcl
# Wait for database to be ready
resource "kubewait_deployments" "database" {
  for       = "condition=Available"
  name      = "postgresql"
  namespace = "database"
  timeout   = 300
}

# Start API server only after database is ready  
resource "kubernetes_deployment" "api" {
  # ... deployment config ...
  depends_on = [kubewait_deployments.database]
}
```

### Job Completion
Wait for initialization jobs to complete:

```hcl
resource "kubewait_jobs" "db_migration" {
  for       = "condition=Complete"
  name      = "migrate-db"
  namespace = "production"
  timeout   = 900
}

resource "kubernetes_deployment" "app" {
  # ... deployment config ...
  depends_on = [kubewait_jobs.db_migration]
}
```

### Multi-Cluster Support
Different resources can target different clusters:

```hcl
# Production cluster
resource "kubewait_nodes" "prod_ready" {
  for              = "condition=Ready"
  all              = true
  kube_config_type = "file"
  kube_config      = "./prod-cluster.kubeconfig"
}

# Staging cluster using raw config
resource "kubewait_nodes" "staging_ready" {
  for              = "condition=Ready"
  all              = true
  kube_config_type = "raw"
  kube_config      = data.aws_eks_cluster.staging.kubeconfig
}
```

## Installation

### Terraform Registry (when published)
```hcl
terraform {
  required_providers {
    kubewait = {
      source  = "nuxij/kubewait"
      version = "~> 1.0"
    }
  }
}
```

### Local Development
1. Clone this repository
2. Build the provider: `go build -o terraform-provider-kubewait`
3. Place in your Terraform plugins directory or use as a local provider

## Provider Configuration

```hcl
provider "kubewait" {
  # Option 1: Use kubeconfig file path
  kube_config_type = "file"
  kube_config      = "~/.kube/config"
  
  # Option 2: Use raw kubeconfig content
  kube_config_type = "raw"
  kube_config      = file("~/.kube/config")
  
  # Option 3: Auto-discovery (default)
  # kube_config_type = "auto"  # Uses in-cluster or ~/.kube/config
  
  # Optional: Specify context
  context = "my-cluster-context"
  
  # Optional: Set default namespace
  namespace = "default"
}
```

The provider supports these authentication types:
- `auto` - Tries in-cluster config, then `~/.kube/config` (default)
- `raw` - Uses raw kubeconfig content from `kube_config`
- `file` - Uses kubeconfig file at path in `kube_config`
- `provider` - Inherits from provider configuration

## Common Resource Schema

All wait resources share these attributes:

### Required
| Attribute | Type | Description |
|-----------|------|-------------|
| `for` | string | Condition to wait for (e.g., "condition=Ready", "phase=Running") |

### Optional
| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `name` | string | - | Specific resource name to wait for |
| `namespace` | string | "default" | Namespace (for namespaced resources) |
| `labels` | string | - | Label selector (e.g., "app=nginx,tier=frontend") |
| `field_selector` | string | - | Field selector (e.g., "spec.nodeName=node1") |
| `all` | bool | false | Wait for ALL matching resources |
| `timeout` | number | 300 | Maximum wait time in seconds |
| `check_interval` | number | 5 | Check interval in seconds |
| `check_once` | bool | false | Only check condition on first apply |
| `kube_config_type` | string | "provider" | Authentication type |
| `kube_config` | string | - | Raw config or file path |
| `context` | string | - | Kubernetes context to use |

### Computed
| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Unique identifier for the wait resource |
| `condition_met` | bool | Whether the condition has been met |
| `last_checked` | string | Timestamp of last condition check |
| `message` | string | Status message about the condition |

## Generic Wait Resource

For resources not covered by specific types:

```hcl
resource "kubewait_wait" "custom" {
  for      = "condition=Ready" 
  resource = "customresources"  # Any K8s resource type
  labels   = "app=my-operator"
  timeout  = 120
}
```

## Supported Condition Types

### For Nodes and Pods
- `condition=Ready` - Resource is ready
- `condition=PodScheduled` - Pod has been scheduled (pods only)
- `condition=Initialized` - Pod has been initialized (pods only)
- `condition=ContainersReady` - All containers are ready (pods only)

### For Pods
- `phase=Running` - Pod is running
- `phase=Succeeded` - Pod has completed successfully  
- `phase=Failed` - Pod has failed

### For Deployments, DaemonSets, StatefulSets
- `condition=Available` - Deployment is available
- `condition=Progressing` - Deployment is progressing

### For Jobs
- `condition=Complete` - Job has completed successfully
- `condition=Failed` - Job has failed

### JSONPath Conditions
For advanced conditions, use JSONPath expressions:
```hcl
# Wait for specific replica count
resource "kubewait_deployments" "app" {
  for  = "jsonpath={.status.readyReplicas}=3"
  name = "my-app"
}

# Wait for load balancer IP
resource "kubewait_services" "lb" {
  for  = "jsonpath={.status.loadBalancer.ingress[0].ip}"
  name = "my-service"
  type = "LoadBalancer"
}
```

## Examples

See the [examples](./examples/) directory for complete usage examples:

- [Basic Usage](./examples/basic-usage.md) - Common patterns and simple examples
- [DigitalOcean Cluster](./examples/digitalocean-cluster.md) - Complete DO cluster setup with Helm

## Building from Source

```bash
# Clone the repository
git clone https://github.com/shadowacre/terraform-provider-kubewait
cd terraform-provider-kubewait

# Download dependencies
go mod tidy

# Build the provider
go build -o terraform-provider-kubewait

# Run tests
go test ./...
```

## Development

This provider is built using the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework). 

### Project Structure

```
.
├── main.go                    # Provider entry point
├── provider/                  # Provider implementation
│   ├── provider.go           # Main provider definition
│   ├── common_resource.go    # Shared resource functionality
│   ├── resource_kubectl_wait.go # Generic wait resource
│   ├── resource_nodes.go     # Node-specific wait resource
│   ├── resource_pods.go      # Pod-specific wait resource
│   └── resource_*.go         # Other resource-specific implementations
├── internal/kubernetes/       # Kubernetes client utilities
│   ├── client.go             # Client configuration and creation
│   └── conditions.go         # Condition checking logic
└── examples/                  # Usage examples
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires running Kubernetes cluster)
go test -tags=integration ./...

# Test with local cluster (kind/minikube)
terraform init
terraform plan
terraform apply
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- HashiCorp for the excellent Terraform Plugin Framework
- The Kubernetes community for the robust client libraries
- Inspiration from `kubectl wait` command functionality