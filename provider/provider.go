package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure KubeWaitProvider satisfies various provider interfaces.
var _ provider.Provider = &KubeWaitProvider{}

// KubeWaitProvider defines the provider implementation.
type KubeWaitProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// KubeWaitProviderModel describes the provider data model.
type KubeWaitProviderModel struct {
	KubeConfigType types.String `tfsdk:"kube_config_type"`
	KubeConfig     types.String `tfsdk:"kube_config"`
	Context        types.String `tfsdk:"context"`
	Namespace      types.String `tfsdk:"namespace"`
}

func (p *KubeWaitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kubewait"
	resp.Version = p.version
}

func (p *KubeWaitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The kube_wait provider allows you to wait for Kubernetes resources to meet specific conditions before proceeding with other Terraform operations.",

		Attributes: map[string]schema.Attribute{
			"kube_config_type": schema.StringAttribute{
				MarkdownDescription: "Type of kube config: 'auto' (default), 'raw', 'file'. If 'auto', uses in-cluster config or ~/.kube/config. If 'raw', uses kube_config content. If 'file', uses kube_config as file path.",
				Optional:            true,
			},
			"kube_config": schema.StringAttribute{
				MarkdownDescription: "Kubernetes config content (when kube_config_type='raw') or file path (when kube_config_type='file')",
				Optional:            true,
				Sensitive:           true,
			},
			"context": schema.StringAttribute{
				MarkdownDescription: "Kubernetes context to use.",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Default namespace for resources.",
				Optional:            true,
			},
		},
	}
}

func (p *KubeWaitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KubeWaitProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// Example client configuration for data sources and resources
	// that require configuration (create, read, update, and delete).

	// Store provider configuration in client for use in resources
	kubeConfigType := data.KubeConfigType.ValueString()
	if kubeConfigType == "" {
		kubeConfigType = "auto"
	}

	providerConfig := &ProviderConfig{
		KubeConfigType: kubeConfigType,
		KubeConfig:     data.KubeConfig.ValueString(),
		Context:        data.Context.ValueString(),
		Namespace:      data.Namespace.ValueString(),
	}

	resp.DataSourceData = providerConfig
	resp.ResourceData = providerConfig
}

func (p *KubeWaitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWaitResource,
		NewNodesResource,
		NewPodsResource,
		NewDeploymentsResource,
		NewDaemonSetsResource,
		NewServicesResource,
		NewStatefulSetsResource,
		NewIngressResource,
		NewJobsResource,
		NewCronJobsResource,
	}
}

func (p *KubeWaitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KubeWaitProvider{
			version: version,
		}
	}
}

// ProviderConfig holds configuration for the provider
type ProviderConfig struct {
	KubeConfigType string
	KubeConfig     string
	Context        string
	Namespace      string
}
