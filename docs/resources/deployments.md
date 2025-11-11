---
page_title: "kubewait_deployments Resource"
description: |-
  Waits for Kubernetes deployments to meet specified conditions.
---

# kubewait_deployments Resource

Waits for Kubernetes deployments to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for deployment to be available
resource "kubewait_deployments" "app_ready" {
  name      = "my-app"
  namespace = "production"
  for       = "condition=Available"
  timeout   = 300
}

# Wait for all deployments with specific label
resource "kubewait_deployments" "all_apps" {
  namespace = "staging"
  labels    = "tier=backend"
  for       = "condition=Available"
  all       = true
  timeout   = 600
}

# Wait for specific replica count using JSONPath
resource "kubewait_deployments" "scaled_app" {
  name      = "web-server"
  namespace = "default"
  for       = "jsonpath={.status.readyReplicas}=5"
  timeout   = 180
}
```

## Schema

### Required

- `for` (String) Condition to wait for (e.g., 'condition=Available', 'condition=Progressing').

### Optional

- `name` (String) Name of a specific deployment to wait for.
- `namespace` (String) Namespace to search for deployments. Defaults to 'default'.
- `labels` (String) Label selector to filter deployments (e.g., 'app=nginx,tier=frontend').
- `field_selector` (String) Field selector to filter deployments.
- `all` (Boolean) Wait for all matching deployments (true) or just one (false). Defaults to false.
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