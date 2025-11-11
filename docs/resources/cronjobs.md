---
page_title: "kubewait_cronjobs Resource"
description: |-
  Waits for Kubernetes CronJobs to meet specified conditions.
---

# kubewait_cronjobs Resource

Waits for Kubernetes CronJobs to meet specified conditions before allowing dependent resources to proceed.

## Example Usage

```terraform
# Wait for cronjobs
resource "kubewait_cronjobs" "example" {
  namespace = "default"
  for       = "jsonpath={.metadata.name}"
  timeout   = 300
}
```

## Schema

### Required

- `for` (String) Condition to wait for.

### Optional

- `name` (String) Name of a specific resource to wait for.
- `namespace` (String) Namespace to search for resources. Defaults to 'default'.
- `labels` (String) Label selector to filter resources.
- `field_selector` (String) Field selector to filter resources.
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
