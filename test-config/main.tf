terraform {
  required_providers {
    kubewait = {
      source  = "nuxij/kubewait"
      version = "~> 1.0"
    }
  }
}

provider "kubewait" {
  # Use auto-discovery for kubeconfig
}

# =============================================================================
# POSITIVE TESTS - Resources that should exist and succeed
# =============================================================================

# 1. Generic kubewait_wait - check that the kubernetes service exists
resource "kubewait_wait" "kubernetes_api_service_check_once" {
  resource       = "services"
  name           = "kubernetes"
  namespace      = "default"
  for            = "jsonpath={.spec.clusterIP}"  # Test jsonpath - just check IP exists
  timeout        = 30
  check_interval = 2
  check_once     = true
}

# 2. kubewait_nodes - verify at least one node is ready
resource "kubewait_nodes" "any_node_ready" {
  for            = "condition=Ready"
  all            = false  # Just need one node ready
  timeout        = 60
  check_interval = 5
}

# 3. kubewait_pods - wait for any kube-system pod to be ready
resource "kubewait_pods" "kube_system_pods" {
  for            = "condition=Ready"
  namespace      = "kube-system"
  all            = false  # Just need one pod ready
  timeout        = 60
  check_interval = 3
}

# 4. kubewait_services - verify kubernetes API service exists
resource "kubewait_services" "kubernetes_api" {
  name           = "kubernetes"
  namespace      = "default"
  for            = "jsonpath={.metadata.name}"  # Just verify service name matches
  timeout        = 30
  check_interval = 2
}

# =============================================================================
# OUTPUT RESULTS
# =============================================================================

output "test_results" {
  value = {
    kubernetes_api_service_check_once = {
      condition_met    = kubewait_wait.kubernetes_api_service_check_once.condition_met
      kube_config_type = kubewait_wait.kubernetes_api_service_check_once.kube_config_type
      message          = kubewait_wait.kubernetes_api_service_check_once.message
    }
    
    any_node_ready = {
      condition_met = kubewait_nodes.any_node_ready.condition_met
      message       = kubewait_nodes.any_node_ready.message
    }
    
    kube_system_pods = {
      condition_met = kubewait_pods.kube_system_pods.condition_met
      message       = kubewait_pods.kube_system_pods.message
    }
    
    kubernetes_api = {
      condition_met = kubewait_services.kubernetes_api.condition_met
      message       = kubewait_services.kubernetes_api.message
    }
  }
}
