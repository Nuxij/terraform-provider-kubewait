package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NodesResource{}

func NewNodesResource() resource.Resource {
	return &NodesResource{}
}

// NodesResource defines the resource implementation.
type NodesResource struct {
	BaseWaitResource
}

// NodesResourceModel describes the resource data model.
type NodesResourceModel = ClusterScopedWaitResourceModel

func (r *NodesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
	r.resourceType = "nodes"
}

func (r *NodesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "nodes",
		Description:      "Waits for Kubernetes nodes to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Ready', 'status=Running')",
		IncludeNamespace: false, // Nodes are cluster-scoped
	})
}

func (r *NodesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NodesResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "nodes"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *NodesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NodesResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "nodes"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *NodesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *NodesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
