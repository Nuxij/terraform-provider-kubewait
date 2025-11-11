---
page_title: "kubewait_nodes Resource"
description: |-
  Waits for Kubernetes nodes to meet specified conditions.
---

# kubewait_nodes Resource

Waits for Kubernetes nodes to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for all nodes to be ready
resource "kubewait_nodes" "cluster_ready" {
  for     = "condition=Ready"
  all     = true
  timeout = 600
}

# Wait for any node to be ready
resource "kubewait_nodes" "any_node" {
  for     = "condition=Ready"
  all     = false
  timeout = 120
}

# Wait for specific node
resource "kubewait_nodes" "master_node" {
  name = "k8s-master-01"
  for  = "condition=Ready"
}
```

## Schema

### Required

- `for` (String) Condition to wait for (e.g., 'condition=Ready', 'status=Running').

### Optional

- `name` (String) Name of a specific node to wait for.
- `labels` (String) Label selector to filter nodes (e.g., 'node-role.kubernetes.io/master').
- `field_selector` (String) Field selector to filter nodes.
- `all` (Boolean) Wait for all matching nodes (true) or just one (false). Defaults to false.
- `timeout` (Number) Maximum time to wait in seconds. Defaults to 300.
- `check_interval` (Number) How often to check the condition in seconds. Defaults to 5.
- `check_once` (Boolean) If true, only check the condition on first apply and skip checks on subsequent plans/applies. Defaults to false.
- `kube_config_type` (String) Type of kube config: 'auto', 'raw', 'file', or 'provider'. If not specified, inherits from provider configuration.
- `kube_config` (String, Sensitive) Kubernetes config content (when kube_config_type='raw') or file path (when kube_config_type='file').
- `context` (String) Kubernetes context to use.

### Read-Only

- `id` (String) Unique identifier for the wait resource.
- `condition_met` (Boolean) Whether the wait condition was met.
- `last_checked` (String) Timestamp of last condition check.
- `message` (String) Status message about the wait condition.