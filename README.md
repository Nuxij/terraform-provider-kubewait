# Terraform Provider: kube_wait

A Terraform provider that allows you to wait for Kubernetes resources to meet specific conditions before proceeding with other Terraform operations. Perfect for ensuring cluster readiness, waiting for deployments, and controlling deployment order in complex Kubernetes environments.

## Features

- ✅ **Multiple Authentication Methods**: Auto-discovery, file paths, raw content, in-cluster
- ✅ **Per-Resource Configuration**: Each resource can authenticate to different clusters
- ✅ **Zero Configuration**: Works out-of-the-box with standard Kubernetes setups
- ✅ Wait for Kubernetes resources to meet conditions (Ready, Available, etc.)
- ✅ Support for nodes, pods, deployments, services, and other resources
- ✅ Label and field selectors for precise resource targeting
- ✅ Configurable timeouts and check intervals
- ✅ Integration with Terraform dependency management (`depends_on`)
- ✅ Context-aware operations

## Use Cases

### DigitalOcean Kubernetes Clusters
Wait for DO's node agent and system pods to be ready before deploying applications:

```hcl
resource "kube_wait" "do-agent-pod" {
  for         = "condition=Ready"
  resource    = "pod"
  labels      = "app=do-node-agent"
  namespace   = "kube-system"
  timeout     = 600
  kube_config = data.digitalocean_kubernetes_cluster.main.kube_config.0.raw_config
}

resource "helm_release" "prometheus" {
  # ... helm configuration ...
  depends_on = [kube_wait.do-agent-pod]
}
```

### Auto-Discovery for Local Development
Works with in-cluster config or ~/.kube/config automatically:

```hcl
resource "kube_wait" "cluster_ready" {
  for      = "condition=Ready"
  resource = "nodes"
  all      = true
  timeout  = 600
  # No auth config needed - uses auto-discovery
}
```

### Deployment Orchestration
Ensure applications start in the correct order:

```hcl
# Wait for database to be ready
resource "kube_wait" "database" {
  for       = "condition=Available"
  resource  = "deployment"
  name      = "postgresql"
  namespace = "database"
  timeout   = 300
}

# Start API server only after database is ready
resource "kubernetes_deployment" "api" {
  # ... deployment config ...
  depends_on = [kube_wait.database]
}
```

### Cluster Initialization
Wait for all nodes and system components:

```hcl
resource "kube_wait" "all-nodes" {
  for      = "condition=Ready"
  resource = "nodes"
  all      = true
  timeout  = 600
  # Uses in-cluster config or ~/.kube/config automatically
}
```

### Multi-Cluster Support
Different resources can target different clusters:

```hcl
# Production cluster
resource "kube_wait" "prod_ready" {
  for              = "condition=Ready"
  resource         = "nodes"
  all              = true
  kube_config_path = "./prod-cluster.kubeconfig"
}

# Staging cluster  
resource "kube_wait" "staging_ready" {
  for         = "condition=Ready"
  resource    = "nodes"
  all         = true
  kube_config = data.aws_eks_cluster.staging.kubeconfig
}
```

## Installation

### Terraform Registry (when published)
```hcl
terraform {
  required_providers {
    kube_wait = {
      source  = "shadowacre/kube_wait"
      version = "~> 1.0"
    }
  }
}
```

### Local Development
1. Clone this repository
2. Build the provider: `go build -o terraform-provider-kube_wait`
3. Place in your Terraform plugins directory or use as a local provider

## Provider Configuration

```hcl
provider "kube_wait" {
  # Option 1: Use kubeconfig file path
  kube_config_path = "~/.kube/config"
  
  # Option 2: Use raw kubeconfig content
  # kube_config = file("~/.kube/config")
  
  # Optional: Specify context
  # context = "my-cluster-context"
  
  # Optional: Set default namespace
  # namespace = "default"
}
```

The provider will automatically try these kubeconfig sources in order:
1. Raw `kube_config` content (if provided)
2. File at `kube_config_path` (if provided) 
3. In-cluster configuration (if running in a pod)
4. `~/.kube/config` (default location)

## Resource: `kube_wait`

### Schema

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `for` | string | ✅ | Condition to wait for (e.g., "condition=Ready", "phase=Running") |
| `resource` | string | ✅ | Kubernetes resource type (nodes, pods, deployments, etc.) |
| `name` | string | ❌ | Specific resource name to wait for |
| `namespace` | string | ❌ | Namespace (uses provider default or "default") |
| `labels` | string | ❌ | Label selector (e.g., "app=nginx,tier=frontend") |
| `field_selector` | string | ❌ | Field selector (e.g., "spec.nodeName=node1") |
| `all` | bool | ❌ | Wait for ALL matching resources (default: false) |
| `timeout` | number | ❌ | Maximum wait time in seconds (default: 300) |
| `check_interval` | number | ❌ | Check interval in seconds (default: 5) |
| `kube_config` | string | ❌ | Raw kubeconfig content (takes precedence) |
| `kube_config_path` | string | ❌ | Path to kubeconfig file |
| `context` | string | ❌ | Kubernetes context to use |

### Computed Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Unique identifier for the wait resource |
| `condition_met` | bool | Whether the condition has been met |
| `last_checked` | string | Timestamp of last condition check |
| `message` | string | Status message about the condition |

### Supported Condition Types

#### For Nodes and Pods
- `condition=Ready` - Resource is ready
- `condition=PodScheduled` - Pod has been scheduled
- `condition=Initialized` - Pod has been initialized
- `condition=ContainersReady` - All containers are ready

#### For Pods Only
- `phase=Running` - Pod is running
- `phase=Succeeded` - Pod has completed successfully
- `phase=Failed` - Pod has failed

#### For Deployments
- `condition=Available` - Deployment is available
- `condition=Progressing` - Deployment is progressing

## Examples

See the [examples](./examples/) directory for complete usage examples:

- [Basic Usage](./examples/basic-usage.md) - Common patterns and simple examples
- [DigitalOcean Cluster](./examples/digitalocean-cluster.md) - Complete DO cluster setup with Helm

## Building from Source

```bash
# Clone the repository
git clone https://github.com/shadowacre/terraform-provider-kube-wait
cd terraform-provider-kube-wait

# Download dependencies
go mod tidy

# Build the provider
go build -o terraform-provider-kube_wait

# Run tests (when implemented)
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
│   └── resource_kubectl_wait.go # kubectl_wait resource
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