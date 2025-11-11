---
page_title: "kubewait_services Resource"
description: |-
  Waits for Kubernetes services to meet specified conditions.
---

# kubewait_services Resource

Waits for Kubernetes services to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for service to have cluster IP assigned
resource "kubewait_services" "api_service" {
  name      = "api-server"
  namespace = "production"
  for       = "jsonpath={.spec.clusterIP}"
  timeout   = 120
}

# Wait for LoadBalancer service to get external IP
resource "kubewait_services" "loadbalancer" {
  name      = "web-lb"
  namespace = "default"
  for       = "jsonpath={.status.loadBalancer.ingress[0].ip}"
  timeout   = 300
}

# Wait for any service with specific labels
resource "kubewait_services" "backend_services" {
  namespace = "backend"
  labels    = "tier=api"
  for       = "jsonpath={.metadata.name}"
  all       = false
}
```

## Schema

### Required

- `for` (String) Condition to wait for (e.g., 'jsonpath={.spec.clusterIP}', 'jsonpath={.status.loadBalancer.ingress[0].ip}').

### Optional

- `name` (String) Name of a specific service to wait for.
- `namespace` (String) Namespace to search for services. Defaults to 'default'.
- `labels` (String) Label selector to filter services (e.g., 'app=nginx,tier=frontend').
- `field_selector` (String) Field selector to filter services (e.g., 'spec.type=LoadBalancer').
- `all` (Boolean) Wait for all matching services (true) or just one (false). Defaults to false.
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