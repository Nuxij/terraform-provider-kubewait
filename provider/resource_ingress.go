package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IngressResource{}

func NewIngressResource() resource.Resource {
	return &IngressResource{}
}

// IngressResource defines the resource implementation.
type IngressResource struct {
	BaseWaitResource
}

// IngressResourceModel describes the resource data model.
type IngressResourceModel = GenericWaitResourceModel

func (r *IngressResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingress"
	r.resourceType = "ingress"
}

func (r *IngressResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "ingress",
		Description:      "Waits for Kubernetes ingress to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'jsonpath={.status.loadBalancer.ingress[0].ip}', 'condition=Ready')",
		IncludeNamespace: true,
	})
}

func (r *IngressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IngressResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "ingress"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *IngressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IngressResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "ingress"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *IngressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *IngressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
