package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DaemonSetsResource{}

func NewDaemonSetsResource() resource.Resource {
	return &DaemonSetsResource{}
}

// DaemonSetsResource defines the resource implementation.
type DaemonSetsResource struct {
	BaseWaitResource
}

// DaemonSetsResourceModel describes the resource data model.
type DaemonSetsResourceModel = GenericWaitResourceModel

func (r *DaemonSetsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daemonsets"
	r.resourceType = "daemonsets"
}

func (r *DaemonSetsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "daemonsets",
		Description:      "Waits for Kubernetes daemonsets to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Available', 'jsonpath={.status.numberReady}')",
		IncludeNamespace: true,
	})
}

func (r *DaemonSetsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DaemonSetsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "daemonsets"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *DaemonSetsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DaemonSetsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "daemonsets"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *DaemonSetsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *DaemonSetsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
