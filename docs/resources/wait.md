---
page_title: "kubewait_wait Resource"
description: |-
  Waits for any Kubernetes resource type to meet specified conditions.
---

# kubewait_wait Resource

Waits for any Kubernetes resource type to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for any Kubernetes resource
resource "kubewait_wait" "custom_resources" {
  resource  = "customresources"
  for       = "condition=Ready"
  labels    = "app=my-operator"
  timeout   = 120
}

# Wait for specific service using field selector
resource "kubewait_wait" "kubernetes_api" {
  resource       = "services"
  field_selector = "metadata.name=kubernetes"
  namespace      = "default"
  for            = "jsonpath={.spec.clusterIP}"
  timeout        = 30
}

# Wait for deployment with JSONPath condition
resource "kubewait_wait" "app_replicas" {
  resource  = "deployments"
  name      = "my-app"
  namespace = "production"
  for       = "jsonpath={.status.readyReplicas}=3"
  timeout   = 300
}
```

## Schema

### Required

- `resource` (String) The Kubernetes resource type to wait for (e.g., 'nodes', 'pods', 'deployments').
- `for` (String) Condition to wait for (e.g., 'condition=Ready', 'condition=Available').

### Optional

- `name` (String) Name of a specific resource to wait for.
- `namespace` (String) Namespace to search for resources. Defaults to 'default' for namespaced resources.
- `labels` (String) Label selector to filter resources (e.g., 'app=nginx,tier=frontend').
- `field_selector` (String) Field selector to filter resources (e.g., 'spec.nodeName=node1').
- `all` (Boolean) Wait for all matching resources (true) or just one (false). Defaults to false.
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

## Import

Import is supported using the following syntax:

```shell
terraform import kubewait_wait.example <resource_type>-wait-<timestamp>
```