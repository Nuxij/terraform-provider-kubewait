package provider

import (
	"context"
	"fmt"
	"time"

	"nuxij/kubewait/internal/kubernetes"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BaseWaitResource contains common functionality for all wait resources
type BaseWaitResource struct {
	providerConfig *ProviderConfig
	resourceType   string
}

// GenericWaitResourceModel extends BaseWaitResourceModel with the resource field for generic waiting
type GenericWaitResourceModel struct {
	// Common wait attributes
	For           types.String `tfsdk:"for"`
	Name          types.String `tfsdk:"name"`
	Namespace     types.String `tfsdk:"namespace"`
	All           types.Bool   `tfsdk:"all"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	CheckInterval types.Int64  `tfsdk:"check_interval"`
	CheckOnce     types.Bool   `tfsdk:"check_once"`
	Labels        types.String `tfsdk:"labels"`
	FieldSelector types.String `tfsdk:"field_selector"`

	// Authentication config
	KubeConfigType types.String `tfsdk:"kube_config_type"`
	KubeConfig     types.String `tfsdk:"kube_config"`
	Context        types.String `tfsdk:"context"`

	// Resource field for generic wait resource (optional, omitted from specific resource schemas)
	Resource types.String `tfsdk:"resource"`

	// Computed attributes
	ID           types.String `tfsdk:"id"`
	ConditionMet types.Bool   `tfsdk:"condition_met"`
	LastChecked  types.String `tfsdk:"last_checked"`
	Message      types.String `tfsdk:"message"`
}

// ClusterScopedWaitResourceModel for cluster-scoped resources (like nodes) that don't have namespaces
type ClusterScopedWaitResourceModel struct {
	// Common wait attributes
	For  types.String `tfsdk:"for"`
	Name types.String `tfsdk:"name"`
	// No namespace field for cluster-scoped resources
	All           types.Bool   `tfsdk:"all"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	CheckInterval types.Int64  `tfsdk:"check_interval"`
	CheckOnce     types.Bool   `tfsdk:"check_once"`
	Labels        types.String `tfsdk:"labels"`
	FieldSelector types.String `tfsdk:"field_selector"`

	// Authentication config
	KubeConfigType types.String `tfsdk:"kube_config_type"`
	KubeConfig     types.String `tfsdk:"kube_config"`
	Context        types.String `tfsdk:"context"`

	// Resource field (auto-populated for specific resources)
	Resource types.String `tfsdk:"resource"`

	// Computed attributes
	ID           types.String `tfsdk:"id"`
	ConditionMet types.Bool   `tfsdk:"condition_met"`
	LastChecked  types.String `tfsdk:"last_checked"`
	Message      types.String `tfsdk:"message"`
}

// ResourceConfig defines resource-specific configuration
type ResourceConfig struct {
	TypeName         string
	Description      string
	ForDescription   string
	IncludeNamespace bool
}

// Configure implements resource.Resource.
func (r *BaseWaitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*ProviderConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

// GetCommonSchema returns the common schema attributes for wait resources
func GetCommonSchema(config ResourceConfig) schema.Schema {
	attributes := map[string]schema.Attribute{
		"for": schema.StringAttribute{
			MarkdownDescription: config.ForDescription,
			Required:            true,
		},
		"all": schema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("Wait for all matching %s (true) or just one (false). Defaults to false.", config.TypeName),
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: "Maximum time to wait in seconds. Defaults to 300.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(300),
		},
		"check_interval": schema.Int64Attribute{
			MarkdownDescription: "How often to check the condition in seconds. Defaults to 5.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(5),
		},
		"check_once": schema.BoolAttribute{
			MarkdownDescription: "If true, only check the condition on first apply and skip checks on subsequent plans/applies. Defaults to false.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"labels": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Label selector to filter %s (e.g., 'app=nginx,tier=frontend')", config.TypeName),
			Optional:            true,
		},
		"field_selector": schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Field selector to filter %s (e.g., 'spec.nodeName=node1')", config.TypeName),
			Optional:            true,
		},

		// Authentication config
		"kube_config_type": schema.StringAttribute{
			MarkdownDescription: "Type of kube config: 'auto', 'raw', 'file', or 'provider'. If not specified, inherits from provider configuration.",
			Optional:            true,
			Computed:            true,
		},
		"kube_config": schema.StringAttribute{
			MarkdownDescription: "Kubernetes config content (when kube_config_type='raw') or file path (when kube_config_type='file')",
			Optional:            true,
			Sensitive:           true,
		},
		"context": schema.StringAttribute{
			MarkdownDescription: "Kubernetes context to use",
			Optional:            true,
		},

		// Computed attributes
		"id": schema.StringAttribute{
			Computed: true,
		},
		"condition_met": schema.BoolAttribute{
			MarkdownDescription: "Whether the wait condition was met",
			Computed:            true,
		},
		"last_checked": schema.StringAttribute{
			MarkdownDescription: "Timestamp of last condition check",
			Computed:            true,
		},
		"message": schema.StringAttribute{
			MarkdownDescription: "Status message about the wait condition",
			Computed:            true,
		},
	}

	// Add namespace and name fields for namespaced resources
	if config.IncludeNamespace {
		attributes["name"] = schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Name of a specific %s to wait for (optional - use labels to select multiple %s)",
				config.TypeName, config.TypeName),
			Optional: true,
		}
		attributes["namespace"] = schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Namespace to search for %s. Defaults to 'default'.", config.TypeName),
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("default"),
		}
	} else {
		attributes["name"] = schema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Name of a specific %s to wait for (optional - use labels to select multiple %s)",
				config.TypeName, config.TypeName),
			Optional: true,
		}
	}

	// Always include resource field - it's auto-populated for specific resources
	attributes["resource"] = schema.StringAttribute{
		MarkdownDescription: "The Kubernetes resource type to wait for (e.g., 'nodes', 'pods', 'deployments'). Auto-populated for specific resources.",
		Optional:            true,
		Computed:            true,
	}

	return schema.Schema{
		MarkdownDescription: config.Description,
		Attributes:          attributes,
	}
}

// GetGenericWaitSchema returns the schema for the generic wait resource with the resource field
func GetGenericWaitSchema() schema.Schema {
	// Start with the common schema for namespaced resources
	baseConfig := ResourceConfig{
		TypeName:         "resources",
		Description:      "Waits for Kubernetes resources to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Ready', 'condition=Available')",
		IncludeNamespace: true,
	}

	baseSchema := GetCommonSchema(baseConfig)

	// Override the resource field to be required for generic wait
	baseSchema.Attributes["resource"] = schema.StringAttribute{
		MarkdownDescription: "The Kubernetes resource type to wait for (e.g., 'nodes', 'pods', 'deployments').",
		Required:            true,
	}

	return baseSchema
}

// Create performs the create operation for wait resources
func (r *BaseWaitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data interface{}) {
	resp.Diagnostics.Append(req.Plan.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract common fields based on the data type
	var (
		forValue, nameValue, labelsValue, fieldSelectorValue string
		namespaceValue                                       string
		allValue                                             bool
		timeoutValue, checkIntervalValue                     int64
		kubeConfigTypeValue, kubeConfigValue, contextValue   string
	)

	// Handle both GenericWaitResourceModel and ClusterScopedWaitResourceModel
	switch d := data.(type) {
	case *GenericWaitResourceModel:
		forValue = d.For.ValueString()
		nameValue = d.Name.ValueString()
		namespaceValue = d.Namespace.ValueString()
		allValue = d.All.ValueBool()
		timeoutValue = d.Timeout.ValueInt64()
		checkIntervalValue = d.CheckInterval.ValueInt64()
		labelsValue = d.Labels.ValueString()
		fieldSelectorValue = d.FieldSelector.ValueString()
		kubeConfigTypeValue = d.KubeConfigType.ValueString()
		kubeConfigValue = d.KubeConfig.ValueString()
		contextValue = d.Context.ValueString()

		// Auto-populate the resource field if we have a resourceType
		if r.resourceType != "" {
			d.Resource = types.StringValue(r.resourceType)
		}

		// Check if resource field is still not set
		if d.Resource.IsNull() || d.Resource.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing resource type",
				"The 'resource' field is required.",
			)
			return
		}

		// Set the internal resource type if not already set
		if r.resourceType == "" {
			r.resourceType = d.Resource.ValueString()
		}

	case *ClusterScopedWaitResourceModel:
		forValue = d.For.ValueString()
		nameValue = d.Name.ValueString()
		namespaceValue = "" // Cluster-scoped resources don't have namespaces
		allValue = d.All.ValueBool()
		timeoutValue = d.Timeout.ValueInt64()
		checkIntervalValue = d.CheckInterval.ValueInt64()
		labelsValue = d.Labels.ValueString()
		fieldSelectorValue = d.FieldSelector.ValueString()
		kubeConfigTypeValue = d.KubeConfigType.ValueString()
		kubeConfigValue = d.KubeConfig.ValueString()
		contextValue = d.Context.ValueString()

		// Auto-populate the resource field if we have a resourceType
		if r.resourceType != "" {
			d.Resource = types.StringValue(r.resourceType)
		}

		// Check if resource field is still not set
		if d.Resource.IsNull() || d.Resource.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing resource type",
				"The 'resource' field is required.",
			)
			return
		}

		// Set the internal resource type if not already set
		if r.resourceType == "" {
			r.resourceType = d.Resource.ValueString()
		}
	default:
		resp.Diagnostics.AddError(
			"Invalid data type",
			"Unsupported resource model type",
		)
		return
	}

	// Create Kubernetes client config
	kubeClientConfig := &kubernetes.ClientConfig{
		KubeConfig:     kubeConfigValue,
		KubeConfigPath: "",
		Context:        contextValue,
	}

	if kubeConfigTypeValue == "" {
		kubeConfigTypeValue = "provider"
	}

	switch kubeConfigTypeValue {
	case "raw":
		kubeClientConfig.KubeConfig = kubeConfigValue
		kubeClientConfig.KubeConfigPath = ""
	case "file":
		kubeClientConfig.KubeConfig = ""
		kubeClientConfig.KubeConfigPath = kubeConfigValue
	case "auto":
		kubeClientConfig.KubeConfig = ""
		kubeClientConfig.KubeConfigPath = ""
	default: // "provider"
		if r.providerConfig != nil {
			switch r.providerConfig.KubeConfigType {
			case "raw":
				kubeClientConfig.KubeConfig = r.providerConfig.KubeConfig
			case "file":
				kubeClientConfig.KubeConfigPath = r.providerConfig.KubeConfig
			default:
				kubeClientConfig.KubeConfig = ""
				kubeClientConfig.KubeConfigPath = ""
			}
			kubeClientConfig.Context = r.providerConfig.Context
		}
	}

	client, err := kubernetes.NewClient(ctx, kubeClientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Kubernetes client",
			err.Error(),
		)
		return
	}

	// Get the correct namespace for the operation
	namespace := r.getNamespaceValue(namespaceValue)

	// Perform the wait operation
	conditionChecker := &kubernetes.ConditionChecker{
		Client: client,
		Config: &kubernetes.WaitConfig{
			Resource:      r.resourceType,
			Name:          nameValue,
			Namespace:     namespace,
			Labels:        labelsValue,
			FieldSelector: fieldSelectorValue,
			Condition:     forValue,
			All:           allValue,
			Timeout:       time.Duration(timeoutValue) * time.Second,
			CheckInterval: time.Duration(checkIntervalValue) * time.Second,
		},
	}

	result, err := conditionChecker.WaitForCondition(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Wait operation failed",
			err.Error(),
		)
		return
	}

	// Set computed values based on data type
	switch d := data.(type) {
	case *GenericWaitResourceModel:
		d.ID = types.StringValue(fmt.Sprintf("%s-wait-%d", r.resourceType, time.Now().Unix()))
		d.ConditionMet = types.BoolValue(result.ConditionMet)
		d.LastChecked = types.StringValue(result.LastChecked.Format(time.RFC3339))
		d.Message = types.StringValue(result.Message)
		if d.KubeConfigType.ValueString() == "" {
			d.KubeConfigType = types.StringValue("provider")
		}
	case *ClusterScopedWaitResourceModel:
		d.ID = types.StringValue(fmt.Sprintf("%s-wait-%d", r.resourceType, time.Now().Unix()))
		d.ConditionMet = types.BoolValue(result.ConditionMet)
		d.LastChecked = types.StringValue(result.LastChecked.Format(time.RFC3339))
		d.Message = types.StringValue(result.Message)
		if d.KubeConfigType.ValueString() == "" {
			d.KubeConfigType = types.StringValue("provider")
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read performs the read operation for wait resources
func (r *BaseWaitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, data interface{}) {
	resp.Diagnostics.Append(req.State.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle resource field - set internal resource type
	switch d := data.(type) {
	case *GenericWaitResourceModel:
		if r.resourceType != "" {
			d.Resource = types.StringValue(r.resourceType)
		} else {
			r.resourceType = d.Resource.ValueString()
		}

		// If check_once is enabled and condition was already met, skip re-checking
		if d.CheckOnce.ValueBool() && d.ConditionMet.ValueBool() {
			resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			return
		}

	case *ClusterScopedWaitResourceModel:
		if r.resourceType != "" {
			d.Resource = types.StringValue(r.resourceType)
		} else {
			r.resourceType = d.Resource.ValueString()
		}

		// If check_once is enabled and condition was already met, skip re-checking
		if d.CheckOnce.ValueBool() && d.ConditionMet.ValueBool() {
			resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			return
		}
	}

	// Re-check the condition with a short timeout
	// ... (implement similar logic as Create but with short timeout for read checks)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update implements resource.Resource for wait resources
func (r *BaseWaitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Wait resources don't support updates - any change should trigger a recreation
	resp.Diagnostics.AddError(
		"Update not supported",
		"Wait resources cannot be updated. Any configuration change requires replacement.",
	)
}

// Delete implements resource.Resource for wait resources
func (r *BaseWaitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to do on delete for wait resources
}

// getNamespaceValue returns the namespace to use, with proper fallback logic
// For cluster-scoped resources, this will return an empty string
func (r *BaseWaitResource) getNamespaceValue(namespaceValue string) string {
	// Cluster-scoped resources (like nodes) don't have namespaces
	if r.resourceType == "nodes" || r.resourceType == "node" ||
		r.resourceType == "persistentvolumes" || r.resourceType == "persistentvolume" ||
		r.resourceType == "clusterroles" || r.resourceType == "clusterrole" ||
		r.resourceType == "clusterrolebindings" || r.resourceType == "clusterrolebinding" {
		return ""
	}

	if namespaceValue != "" {
		return namespaceValue
	}
	if r.providerConfig != nil && r.providerConfig.Namespace != "" {
		return r.providerConfig.Namespace
	}
	return "default"
}

// getNamespace returns the namespace to use, with proper fallback logic
// For cluster-scoped resources, this will return an empty string
func (r *BaseWaitResource) getNamespace(data GenericWaitResourceModel) string {
	return r.getNamespaceValue(data.Namespace.ValueString())
}

// getKubeClientConfig creates a Kubernetes client config from resource and provider settings
func (r *BaseWaitResource) getKubeClientConfig(data GenericWaitResourceModel) *kubernetes.ClientConfig {
	configType := data.KubeConfigType.ValueString()
	if configType == "" {
		configType = "provider"
	}

	switch configType {
	case "raw":
		return &kubernetes.ClientConfig{
			KubeConfig:     data.KubeConfig.ValueString(),
			KubeConfigPath: "",
			Context:        data.Context.ValueString(),
		}
	case "file":
		return &kubernetes.ClientConfig{
			KubeConfig:     "",
			KubeConfigPath: data.KubeConfig.ValueString(),
			Context:        data.Context.ValueString(),
		}
	case "auto":
		return &kubernetes.ClientConfig{
			KubeConfig:     "",
			KubeConfigPath: "",
			Context:        data.Context.ValueString(),
		}
	default: // "provider" or any other value
		if r.providerConfig != nil {
			// Determine provider config type
			providerConfigType := r.providerConfig.KubeConfigType
			if providerConfigType == "" {
				providerConfigType = "auto"
			}

			switch providerConfigType {
			case "raw":
				return &kubernetes.ClientConfig{
					KubeConfig:     r.providerConfig.KubeConfig,
					KubeConfigPath: "",
					Context:        r.providerConfig.Context,
				}
			case "file":
				return &kubernetes.ClientConfig{
					KubeConfig:     "",
					KubeConfigPath: r.providerConfig.KubeConfig,
					Context:        r.providerConfig.Context,
				}
			default: // "auto"
				return &kubernetes.ClientConfig{
					KubeConfig:     "",
					KubeConfigPath: "",
					Context:        r.providerConfig.Context,
				}
			}
		}
		// Fallback to auto-discovery if no provider config
		return &kubernetes.ClientConfig{
			KubeConfig:     "",
			KubeConfigPath: "",
			Context:        "",
		}
	}
}
