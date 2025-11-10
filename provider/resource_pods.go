package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PodsResource{}

func NewPodsResource() resource.Resource {
	return &PodsResource{}
}

// PodsResource defines the resource implementation.
type PodsResource struct {
	BaseWaitResource
}

// PodsResourceModel describes the resource data model.
type PodsResourceModel = GenericWaitResourceModel

func (r *PodsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pods"
	r.resourceType = "pods"
}

func (r *PodsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "pods",
		Description:      "Waits for Kubernetes pods to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Ready', 'condition=PodScheduled')",
		IncludeNamespace: true,
	})
}

func (r *PodsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PodsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "pods"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *PodsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PodsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "pods"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *PodsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *PodsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
