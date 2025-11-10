package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServicesResource{}

func NewServicesResource() resource.Resource {
	return &ServicesResource{}
}

// ServicesResource defines the resource implementation.
type ServicesResource struct {
	BaseWaitResource
}

// ServicesResourceModel describes the resource data model.
type ServicesResourceModel = GenericWaitResourceModel

func (r *ServicesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_services"
	r.resourceType = "services"
}

func (r *ServicesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "services",
		Description:      "Waits for Kubernetes services to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'jsonpath={.status.loadBalancer.ingress[0].ip}', 'jsonpath={.spec.clusterIP}')",
		IncludeNamespace: true,
	})
}

func (r *ServicesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServicesResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "services"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *ServicesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServicesResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "services"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *ServicesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *ServicesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
