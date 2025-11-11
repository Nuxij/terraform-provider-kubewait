---
page_title: "kubewait_pods Resource"
description: |-
  Waits for Kubernetes pods to meet specified conditions.
---

# kubewait_pods Resource

Waits for Kubernetes pods to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for any pod in kube-system to be ready
resource "kubewait_pods" "system_pods" {
  namespace = "kube-system"
  for       = "condition=Ready"
  all       = false
  timeout   = 120
}

# Wait for specific application pods
resource "kubewait_pods" "app_pods" {
  namespace = "production"
  labels    = "app=my-app,version=v1.0"
  for       = "condition=Ready"
  all       = true
  timeout   = 300
}

# Wait for pod to reach running phase
resource "kubewait_pods" "running_pod" {
  name      = "my-pod"
  namespace = "default"
  for       = "phase=Running"
}
```

## Schema

### Required

- `for` (String) Condition to wait for (e.g., 'condition=Ready', 'condition=PodScheduled', 'phase=Running').

### Optional

- `name` (String) Name of a specific pod to wait for.
- `namespace` (String) Namespace to search for pods. Defaults to 'default'.
- `labels` (String) Label selector to filter pods (e.g., 'app=nginx,tier=frontend').
- `field_selector` (String) Field selector to filter pods (e.g., 'spec.nodeName=node1').
- `all` (Boolean) Wait for all matching pods (true) or just one (false). Defaults to false.
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