package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DeploymentsResource{}

func NewDeploymentsResource() resource.Resource {
	return &DeploymentsResource{}
}

// DeploymentsResource defines the resource implementation.
type DeploymentsResource struct {
	BaseWaitResource
}

// DeploymentsResourceModel describes the resource data model.
type DeploymentsResourceModel = GenericWaitResourceModel

func (r *DeploymentsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
	r.resourceType = "deployments"
}

func (r *DeploymentsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "deployments",
		Description:      "Waits for Kubernetes deployments to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Available', 'condition=Progressing')",
		IncludeNamespace: true,
	})
}

func (r *DeploymentsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "deployments"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *DeploymentsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "deployments"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *DeploymentsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *DeploymentsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
