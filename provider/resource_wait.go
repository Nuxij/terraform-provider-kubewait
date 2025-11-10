package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &WaitResource{}
var _ resource.ResourceWithImportState = &WaitResource{}

func NewWaitResource() resource.Resource {
	return &WaitResource{}
}

// WaitResource defines the resource implementation.
type WaitResource struct {
	BaseWaitResource
}

// WaitResourceModel describes the resource data model.
type WaitResourceModel = GenericWaitResourceModel

func (r *WaitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wait"
}

func (r *WaitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetGenericWaitSchema()
}

func (r *WaitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WaitResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the resource type from the plan
	r.resourceType = data.Resource.ValueString()

	// Use the base wait resource functionality
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *WaitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WaitResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the resource type from the state
	r.resourceType = data.Resource.ValueString()

	// Use the base wait resource functionality
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *WaitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *WaitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}

func (r *WaitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
