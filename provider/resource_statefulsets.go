package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StatefulSetsResource{}

func NewStatefulSetsResource() resource.Resource {
	return &StatefulSetsResource{}
}

// StatefulSetsResource defines the resource implementation.
type StatefulSetsResource struct {
	BaseWaitResource
}

// StatefulSetsResourceModel describes the resource data model.
type StatefulSetsResourceModel = GenericWaitResourceModel

func (r *StatefulSetsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statefulsets"
	r.resourceType = "statefulsets"
}

func (r *StatefulSetsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "statefulsets",
		Description:      "Waits for Kubernetes statefulsets to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'jsonpath={.status.readyReplicas}=3', 'jsonpath={.status.replicas}={.status.readyReplicas}')",
		IncludeNamespace: true,
	})
}

func (r *StatefulSetsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StatefulSetsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "statefulsets"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *StatefulSetsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatefulSetsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "statefulsets"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *StatefulSetsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *StatefulSetsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
