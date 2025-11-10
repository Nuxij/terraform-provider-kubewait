package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CronJobsResource{}

func NewCronJobsResource() resource.Resource {
	return &CronJobsResource{}
}

// CronJobsResource defines the resource implementation.
type CronJobsResource struct {
	BaseWaitResource
}

// CronJobsResourceModel describes the resource data model.
type CronJobsResourceModel = GenericWaitResourceModel

func (r *CronJobsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cronjobs"
	r.resourceType = "cronjobs"
}

func (r *CronJobsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetCommonSchema(ResourceConfig{
		TypeName:         "cronjobs",
		Description:      "Waits for Kubernetes cronjobs to meet specified conditions before allowing dependent resources to proceed.",
		ForDescription:   "Condition to wait for (e.g., 'condition=Complete', 'condition=Failed')",
		IncludeNamespace: true,
	})
}

func (r *CronJobsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CronJobsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "cronjobs"
	r.BaseWaitResource.Create(ctx, req, resp, &data)
}

func (r *CronJobsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CronJobsResourceModel
	// Set the resource type before calling the base method
	r.resourceType = "cronjobs"
	r.BaseWaitResource.Read(ctx, req, resp, &data)
}

func (r *CronJobsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.BaseWaitResource.Update(ctx, req, resp)
}

func (r *CronJobsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.BaseWaitResource.Delete(ctx, req, resp)
}
