package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// WaitConfig holds the configuration for waiting on Kubernetes resources
type WaitConfig struct {
	Resource      string        // Resource type (e.g., "nodes", "pods", "deployments")
	Name          string        // Specific resource name (optional)
	Namespace     string        // Namespace (for namespaced resources)
	Labels        string        // Label selector
	FieldSelector string        // Field selector
	Condition     string        // Condition to wait for (e.g., "condition=Ready")
	All           bool          // Wait for all matching resources
	Timeout       time.Duration // Maximum wait time
	CheckInterval time.Duration // Interval between checks
}

// WaitResult holds the result of a wait operation
type WaitResult struct {
	ConditionMet bool      // Whether the condition was met
	LastChecked  time.Time // When the condition was last checked
	Message      string    // Status message
}

// ConditionChecker provides functionality to wait for Kubernetes resource conditions
type ConditionChecker struct {
	Client *Client
	Config *WaitConfig
}

// conditionCheckFunc defines the signature for condition checking functions
type conditionCheckFunc func(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error)

// resourceConditionCheckers maps resource types to their condition checking functions
var resourceConditionCheckers = map[string]conditionCheckFunc{}

// WaitForCondition waits for the specified condition to be met
func (c *ConditionChecker) WaitForCondition(ctx context.Context) (*WaitResult, error) {
	deadline := time.Now().Add(c.Config.Timeout)
	ticker := time.NewTicker(c.Config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return &WaitResult{
				ConditionMet: false,
				LastChecked:  time.Now(),
				Message:      "Context cancelled",
			}, ctx.Err()

		case <-time.After(time.Until(deadline)):
			return &WaitResult{
				ConditionMet: false,
				LastChecked:  time.Now(),
				Message:      fmt.Sprintf("Timeout after %v", c.Config.Timeout),
			}, fmt.Errorf("timeout waiting for condition %s", c.Config.Condition)

		case <-ticker.C:
			result, err := c.CheckCondition(ctx)
			if err != nil {
				return result, err
			}

			if result.ConditionMet {
				return result, nil
			}

			// Continue waiting if condition not met
		}
	}
}

// CheckCondition performs a single check of the condition
func (c *ConditionChecker) CheckCondition(ctx context.Context) (*WaitResult, error) {
	now := time.Now()

	// Initialize the resource condition checkers map if empty
	if len(resourceConditionCheckers) == 0 {
		c.initResourceCheckers()
	}

	// Parse the condition
	conditionType, conditionValue, err := c.parseCondition()
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Invalid condition format: %s", err),
		}, err
	}

	// Get the appropriate condition checker function
	resourceType := strings.ToLower(c.Config.Resource)
	if checkFunc, exists := resourceConditionCheckers[resourceType]; exists {
		return checkFunc(ctx, conditionType, conditionValue)
	}

	// Handle plural forms by trying to remove 's'
	if strings.HasSuffix(resourceType, "s") {
		singularType := strings.TrimSuffix(resourceType, "s")
		if checkFunc, exists := resourceConditionCheckers[singularType]; exists {
			return checkFunc(ctx, conditionType, conditionValue)
		}
	}

	// Fallback to generic condition checking
	return c.checkGenericCondition(ctx, conditionType, conditionValue)
}

// initResourceCheckers initializes the map of resource condition checkers
func (c *ConditionChecker) initResourceCheckers() {
	resourceConditionCheckers = map[string]conditionCheckFunc{
		"node":         c.checkNodeCondition,
		"nodes":        c.checkNodeCondition,
		"pod":          c.checkPodCondition,
		"pods":         c.checkPodCondition,
		"deployment":   c.checkDeploymentCondition,
		"deployments":  c.checkDeploymentCondition,
		"service":      c.checkServiceCondition,
		"services":     c.checkServiceCondition,
		"daemonset":    c.checkDaemonSetCondition,
		"daemonsets":   c.checkDaemonSetCondition,
		"statefulset":  c.checkStatefulSetCondition,
		"statefulsets": c.checkStatefulSetCondition,
		"job":          c.checkJobCondition,
		"jobs":         c.checkJobCondition,
		"cronjob":      c.checkCronJobCondition,
		"cronjobs":     c.checkCronJobCondition,
		"ingress":      c.checkIngressCondition,
	}
}

// parseCondition parses condition strings like "condition=Ready" or "jsonpath=.status.phase==Running"
func (c *ConditionChecker) parseCondition() (string, string, error) {
	parts := strings.SplitN(c.Config.Condition, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("condition must be in format 'type=value'")
	}

	conditionType := strings.ToLower(strings.TrimSpace(parts[0]))
	conditionValue := strings.TrimSpace(parts[1])

	return conditionType, conditionValue, nil
}

// checkNodeCondition checks conditions on nodes
func (c *ConditionChecker) checkNodeCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	nodeList, err := c.Client.Clientset.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list nodes: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific node name
		filteredNodes := []corev1.Node{}
		for _, node := range nodeList.Items {
			if node.Name == c.Config.Name {
				filteredNodes = append(filteredNodes, node)
			}
		}
		nodeList.Items = filteredNodes
	}

	if len(nodeList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching nodes found",
		}, nil
	}

	readyNodes := 0
	totalNodes := len(nodeList.Items)

	for _, node := range nodeList.Items {
		if conditionType == "condition" {
			for _, condition := range node.Status.Conditions {
				if string(condition.Type) == conditionValue && condition.Status == corev1.ConditionTrue {
					readyNodes++
					break
				}
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyNodes == totalNodes
		message = fmt.Sprintf("%d/%d nodes meet condition %s", readyNodes, totalNodes, c.Config.Condition)
	} else {
		conditionMet = readyNodes > 0
		message = fmt.Sprintf("%d/%d nodes meet condition %s", readyNodes, totalNodes, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkPodCondition checks conditions on pods
func (c *ConditionChecker) checkPodCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	podList, err := c.Client.Clientset.CoreV1().Pods(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list pods: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific pod name
		filteredPods := []corev1.Pod{}
		for _, pod := range podList.Items {
			if pod.Name == c.Config.Name {
				filteredPods = append(filteredPods, pod)
			}
		}
		podList.Items = filteredPods
	}

	if len(podList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching pods found",
		}, nil
	}

	readyPods := 0
	totalPods := len(podList.Items)

	for _, pod := range podList.Items {
		if conditionType == "condition" {
			for _, condition := range pod.Status.Conditions {
				if string(condition.Type) == conditionValue && condition.Status == corev1.ConditionTrue {
					readyPods++
					break
				}
			}
		} else if conditionType == "phase" {
			if string(pod.Status.Phase) == conditionValue {
				readyPods++
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyPods == totalPods
		message = fmt.Sprintf("%d/%d pods meet condition %s", readyPods, totalPods, c.Config.Condition)
	} else {
		conditionMet = readyPods > 0
		message = fmt.Sprintf("%d/%d pods meet condition %s", readyPods, totalPods, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkDeploymentCondition checks conditions on deployments
func (c *ConditionChecker) checkDeploymentCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	deploymentList, err := c.Client.Clientset.AppsV1().Deployments(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list deployments: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific deployment name
		foundDeployment := false
		for _, deployment := range deploymentList.Items {
			if deployment.Name == c.Config.Name {
				foundDeployment = true
				break
			}
		}
		if !foundDeployment {
			deploymentList.Items = nil
		}
	}

	if len(deploymentList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching deployments found",
		}, nil
	}

	readyDeployments := 0
	totalDeployments := len(deploymentList.Items)

	for _, deployment := range deploymentList.Items {
		if conditionType == "condition" {
			for _, condition := range deployment.Status.Conditions {
				if string(condition.Type) == conditionValue && string(condition.Status) == "True" {
					readyDeployments++
					break
				}
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyDeployments == totalDeployments
		message = fmt.Sprintf("%d/%d deployments meet condition %s", readyDeployments, totalDeployments, c.Config.Condition)
	} else {
		conditionMet = readyDeployments > 0
		message = fmt.Sprintf("%d/%d deployments meet condition %s", readyDeployments, totalDeployments, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkServiceCondition checks conditions on services
func (c *ConditionChecker) checkServiceCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	serviceList, err := c.Client.Clientset.CoreV1().Services(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list services: %s", err),
		}, err
	}

	if len(serviceList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching services found",
		}, nil
	}

	// Services don't typically have conditions like pods/nodes,
	// so we'll check for existence or other properties
	return &WaitResult{
		ConditionMet: true,
		LastChecked:  now,
		Message:      fmt.Sprintf("Found %d matching services", len(serviceList.Items)),
	}, nil
}

// checkGenericCondition checks conditions on arbitrary resources using dynamic client
func (c *ConditionChecker) checkGenericCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	// This is a simplified implementation
	// In a full implementation, you would use the dynamic client to query arbitrary resources
	// and apply JSONPath expressions or other condition checks

	dynamicClient, err := dynamic.NewForConfig(c.Client.Config)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to create dynamic client: %s", err),
		}, err
	}

	// For now, we'll return a placeholder implementation
	// In practice, you would need to:
	// 1. Parse the resource type to get GroupVersionResource
	// 2. Use dynamicClient.Resource(gvr).Namespace(namespace).List()
	// 3. Apply the condition logic to the returned unstructured objects

	_ = dynamicClient // Prevent unused variable error

	return &WaitResult{
		ConditionMet: false,
		LastChecked:  now,
		Message:      fmt.Sprintf("Generic condition checking for resource type '%s' not yet implemented", c.Config.Resource),
	}, fmt.Errorf("generic condition checking not implemented for resource type: %s", c.Config.Resource)
}

// checkDaemonSetCondition checks conditions on daemonsets
func (c *ConditionChecker) checkDaemonSetCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	daemonsetList, err := c.Client.Clientset.AppsV1().DaemonSets(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list daemonsets: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific daemonset name
		filteredDS := []interface{}{}
		for _, ds := range daemonsetList.Items {
			if ds.Name == c.Config.Name {
				filteredDS = append(filteredDS, ds)
			}
		}
		if len(filteredDS) == 0 {
			daemonsetList.Items = nil
		}
	}

	if len(daemonsetList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching daemonsets found",
		}, nil
	}

	readyDS := 0
	totalDS := len(daemonsetList.Items)

	for _, ds := range daemonsetList.Items {
		if conditionType == "condition" {
			// DaemonSets don't have standard conditions, check if desired pods are ready
			if ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
				readyDS++
			}
		} else if conditionType == "ready" {
			if ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
				readyDS++
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyDS == totalDS
		message = fmt.Sprintf("%d/%d daemonsets meet condition %s", readyDS, totalDS, c.Config.Condition)
	} else {
		conditionMet = readyDS > 0
		message = fmt.Sprintf("%d/%d daemonsets meet condition %s", readyDS, totalDS, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkStatefulSetCondition checks conditions on statefulsets
func (c *ConditionChecker) checkStatefulSetCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	statefulsetList, err := c.Client.Clientset.AppsV1().StatefulSets(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list statefulsets: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific statefulset name
		filteredSS := []interface{}{}
		for _, ss := range statefulsetList.Items {
			if ss.Name == c.Config.Name {
				filteredSS = append(filteredSS, ss)
			}
		}
		if len(filteredSS) == 0 {
			statefulsetList.Items = nil
		}
	}

	if len(statefulsetList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching statefulsets found",
		}, nil
	}

	readySS := 0
	totalSS := len(statefulsetList.Items)

	for _, ss := range statefulsetList.Items {
		if conditionType == "condition" || conditionType == "ready" {
			// Check if all replicas are ready
			if ss.Spec.Replicas != nil && ss.Status.ReadyReplicas == *ss.Spec.Replicas {
				readySS++
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readySS == totalSS
		message = fmt.Sprintf("%d/%d statefulsets meet condition %s", readySS, totalSS, c.Config.Condition)
	} else {
		conditionMet = readySS > 0
		message = fmt.Sprintf("%d/%d statefulsets meet condition %s", readySS, totalSS, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkJobCondition checks conditions on jobs
func (c *ConditionChecker) checkJobCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	jobList, err := c.Client.Clientset.BatchV1().Jobs(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list jobs: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific job name
		filteredJobs := []interface{}{}
		for _, job := range jobList.Items {
			if job.Name == c.Config.Name {
				filteredJobs = append(filteredJobs, job)
			}
		}
		if len(filteredJobs) == 0 {
			jobList.Items = nil
		}
	}

	if len(jobList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching jobs found",
		}, nil
	}

	readyJobs := 0
	totalJobs := len(jobList.Items)

	for _, job := range jobList.Items {
		if conditionType == "condition" {
			for _, condition := range job.Status.Conditions {
				if string(condition.Type) == conditionValue && string(condition.Status) == "True" {
					readyJobs++
					break
				}
			}
		} else if conditionType == "complete" {
			if job.Status.Succeeded > 0 {
				readyJobs++
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyJobs == totalJobs
		message = fmt.Sprintf("%d/%d jobs meet condition %s", readyJobs, totalJobs, c.Config.Condition)
	} else {
		conditionMet = readyJobs > 0
		message = fmt.Sprintf("%d/%d jobs meet condition %s", readyJobs, totalJobs, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkCronJobCondition checks conditions on cronjobs
func (c *ConditionChecker) checkCronJobCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	cronjobList, err := c.Client.Clientset.BatchV1().CronJobs(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list cronjobs: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific cronjob name
		filteredCJs := []interface{}{}
		for _, cj := range cronjobList.Items {
			if cj.Name == c.Config.Name {
				filteredCJs = append(filteredCJs, cj)
			}
		}
		if len(filteredCJs) == 0 {
			cronjobList.Items = nil
		}
	}

	if len(cronjobList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching cronjobs found",
		}, nil
	}

	// For cronjobs, we can check existence or JSONPath expressions
	readyCJs := 0
	totalCJs := len(cronjobList.Items)

	for range cronjobList.Items {
		if conditionType == "jsonpath" {
			// For JSONPath conditions like jsonpath={.metadata.name}, just check existence
			readyCJs++
		} else if conditionType == "exist" || conditionType == "exists" {
			readyCJs++
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyCJs == totalCJs
		message = fmt.Sprintf("%d/%d cronjobs meet condition %s", readyCJs, totalCJs, c.Config.Condition)
	} else {
		conditionMet = readyCJs > 0
		message = fmt.Sprintf("%d/%d cronjobs meet condition %s", readyCJs, totalCJs, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}

// checkIngressCondition checks conditions on ingress resources
func (c *ConditionChecker) checkIngressCondition(ctx context.Context, conditionType, conditionValue string) (*WaitResult, error) {
	now := time.Now()

	listOptions := metav1.ListOptions{}
	if c.Config.Labels != "" {
		listOptions.LabelSelector = c.Config.Labels
	}
	if c.Config.FieldSelector != "" {
		listOptions.FieldSelector = c.Config.FieldSelector
	}

	ingressList, err := c.Client.Clientset.NetworkingV1().Ingresses(c.Config.Namespace).List(ctx, listOptions)
	if err != nil {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      fmt.Sprintf("Failed to list ingresses: %s", err),
		}, err
	}

	if c.Config.Name != "" {
		// Filter by specific ingress name
		filteredIngresses := []interface{}{}
		for _, ing := range ingressList.Items {
			if ing.Name == c.Config.Name {
				filteredIngresses = append(filteredIngresses, ing)
			}
		}
		if len(filteredIngresses) == 0 {
			ingressList.Items = nil
		}
	}

	if len(ingressList.Items) == 0 {
		return &WaitResult{
			ConditionMet: false,
			LastChecked:  now,
			Message:      "No matching ingresses found",
		}, nil
	}

	// For ingress, we can check existence or load balancer IP assignment
	readyIngresses := 0
	totalIngresses := len(ingressList.Items)

	for _, ing := range ingressList.Items {
		if conditionType == "jsonpath" {
			// For JSONPath conditions like jsonpath={.metadata.name}, just check existence
			readyIngresses++
		} else if conditionType == "exist" || conditionType == "exists" {
			readyIngresses++
		} else if conditionType == "loadbalancer" {
			// Check if load balancer has been assigned
			if len(ing.Status.LoadBalancer.Ingress) > 0 {
				readyIngresses++
			}
		}
	}

	conditionMet := false
	var message string

	if c.Config.All {
		conditionMet = readyIngresses == totalIngresses
		message = fmt.Sprintf("%d/%d ingresses meet condition %s", readyIngresses, totalIngresses, c.Config.Condition)
	} else {
		conditionMet = readyIngresses > 0
		message = fmt.Sprintf("%d/%d ingresses meet condition %s", readyIngresses, totalIngresses, c.Config.Condition)
	}

	return &WaitResult{
		ConditionMet: conditionMet,
		LastChecked:  now,
		Message:      message,
	}, nil
}
