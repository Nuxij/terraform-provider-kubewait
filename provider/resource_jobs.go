package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &JobsResource{}

func NewJobsResource() resource.Resource {
	return &JobsResource{}
}

// JobsResource defines the resource implementation.
type JobsResource struct {
	BaseWaitResource
}

// JobsResourceModel describes the resource data model.
type JobsResourceModel = GenericWaitResourceModel

func (r *JobsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jobs"
	r.resourceType = "jobs"
}

func (r *JobsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "jobs",
		Description:      "Waits for Kubernetes jobs to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Complete', 'condition=Failed')",
		IncludeNamespace: true,
	})
}

func (r *JobsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JobsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "jobs"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *JobsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JobsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "jobs"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *JobsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *JobsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
