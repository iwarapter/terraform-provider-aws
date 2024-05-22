// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/fms/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_fms_delegated_administrator", name="Delegated Administrator")
func newResourceDelegatedAdministrator(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDelegatedAdministrator{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDelegatedAdministrator = "Delegated Administrator"
)

type resourceDelegatedAdministrator struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDelegatedAdministrator) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_fms_delegated_administrator"
}

func (r *resourceDelegatedAdministrator) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"admin_account": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_scope": schema.SingleNestedAttribute{
				Required:   true,
				Attributes: adminScopeAttributes(),
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func adminScopeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"account_scope": schema.SingleNestedAttribute{
			Attributes: accountScopeAttributes(),
		},
		"organization_unit_scope": schema.SingleNestedAttribute{
			Attributes: organizationUnitScopeAttributes(),
		},
		"policy_type_scope": schema.SingleNestedAttribute{
			Attributes: policyTypeScopeAttributes(),
		},
		"region_scope": schema.SingleNestedAttribute{
			Attributes: regionScopeAttributes(),
		},
	}
}

func accountScopeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"accounts": schema.ListAttribute{
			ElementType: types.StringType,
			Required:    true,
		},
		"all_accounts_enabled": schema.BoolAttribute{
			Optional: true,
		},
		"exclude_specified_accounts": schema.BoolAttribute{
			Optional: true,
		},
	}
}

func organizationUnitScopeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"organization_units": schema.ListAttribute{
			ElementType: types.StringType,
		},
		"all_organization_units_enabled": schema.BoolAttribute{
			Optional: true,
		},
		"exclude_specified_organization_units": schema.BoolAttribute{
			Optional: true,
		},
	}
}

func policyTypeScopeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"policy_types": schema.ListAttribute{
			ElementType: types.StringType,
		},
		"all_policy_types_enabled": schema.BoolAttribute{
			Optional: true,
		},
	}
}

func regionScopeAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"regions": schema.ListAttribute{
			ElementType: types.StringType,
		},
		"all_regions_enabled": schema.BoolAttribute{
			Optional: true,
		},
	}
}

func (r *resourceDelegatedAdministrator) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().FMSClient(ctx)

	var plan resourceDelegatedAdministratorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &fms.PutAdminAccountInput{
		AdminAccount: aws.String(plan.AdminAccount.ValueString()),
		AdminScope:   expandAdminScope(ctx, plan.AdminScope),
	}

	out, err := conn.PutAdminAccount(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionCreating, ResNameDelegatedAdministrator, plan.AdminAccount.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionCreating, ResNameDelegatedAdministrator, plan.AdminAccount.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = plan.AdminAccount

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDelegatedAdministratorCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForCreation, ResNameDelegatedAdministrator, plan.AdminAccount.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDelegatedAdministrator) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FMSClient(ctx)

	var state resourceDelegatedAdministratorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDelegatedAdministratorByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionSetting, ResNameDelegatedAdministrator, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, state.AdminAccount.ValueStringPointer())
	state.AdminAccount = flex.StringToFramework(ctx, state.AdminAccount.ValueStringPointer())
	state.Status = flex.StringToFramework(ctx, aws.String(string(out.Status)))
	state.AdminScope = flattenAdminScope(ctx, out.AdminScope)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDelegatedAdministrator) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().FMSClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceDelegatedAdministratorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if plan.AdminScope != state.AdminScope {

		in := &fms.PutAdminAccountInput{
			AdminAccount: aws.String(plan.AdminAccount.ValueString()),
			AdminScope:   expandAdminScope(ctx, plan.AdminScope),
		}

		out, err := conn.PutAdminAccount(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FMS, create.ErrActionUpdating, ResNameDelegatedAdministrator, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FMS, create.ErrActionUpdating, ResNameDelegatedAdministrator, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ID = flex.StringToFramework(ctx, state.AdminAccount.ValueStringPointer())
		plan.AdminAccount = flex.StringToFramework(ctx, state.AdminAccount.ValueStringPointer())

	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	out, err := waitDelegatedAdministratorUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForUpdate, ResNameDelegatedAdministrator, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	plan.Status = flex.StringToFramework(ctx, aws.String(string(out.Status)))
	plan.AdminScope = flattenAdminScope(ctx, out.AdminScope)

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDelegatedAdministrator) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//// TIP: ==== RESOURCE DELETE ====
	//// Most resources have Delete functions. There are rare situations
	//// where you might not need a delete:
	//// a. The AWS API does not provide a way to delete the resource
	//// b. The point of your resource is to perform an action (e.g., reboot a
	////    server) and deleting serves no purpose.
	////
	//// The Delete function should do the following things. Make sure there
	//// is a good reason if you don't do one of these.
	////
	//// 1. Get a client connection to the relevant service
	//// 2. Fetch the state
	//// 3. Populate a delete input structure
	//// 4. Call the AWS delete function
	//// 5. Use a waiter to wait for delete to complete
	//// TIP: -- 1. Get a client connection to the relevant service
	//conn := r.Meta().FMSClient(ctx)
	//
	//// TIP: -- 2. Fetch the state
	//var state resourceDelegatedAdministratorData
	//resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//// TIP: -- 3. Populate a delete input structure
	//in := &fms.DeleteDelegatedAdministratorInput{
	//	DelegatedAdministratorId: aws.String(state.ID.ValueString()),
	//}
	//
	//// TIP: -- 4. Call the AWS delete function
	//_, err := conn.DeleteDelegatedAdministrator(ctx, in)
	//// TIP: On rare occassions, the API returns a not found error after deleting a
	//// resource. If that happens, we don't want it to show up as an error.
	//if err != nil {
	//	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
	//		return
	//	}
	//	resp.Diagnostics.AddError(
	//		create.ProblemStandardMessage(names.FMS, create.ErrActionDeleting, ResNameDelegatedAdministrator, state.ID.String(), err),
	//		err.Error(),
	//	)
	//	return
	//}
	//
	//// TIP: -- 5. Use a waiter to wait for delete to complete
	//deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	//_, err = waitDelegatedAdministratorDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForDeletion, ResNameDelegatedAdministrator, state.ID.String(), err),
	//		err.Error(),
	//	)
	//	return
	//}

}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceDelegatedAdministrator) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitDelegatedAdministratorCreated(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.PutAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.OrganizationStatusOnboarding)},
		Target:                    []string{string(awstypes.OrganizationStatusOnboardingComplete)},
		Refresh:                   statusDelegatedAdministrator(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.PutAdminAccountOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDelegatedAdministratorUpdated(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.GetAdminScopeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.OrganizationStatusOnboarding)},
		Target:                    []string{string(awstypes.OrganizationStatusOnboardingComplete)},
		Refresh:                   statusDelegatedAdministrator(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.GetAdminScopeOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitDelegatedAdministratorDeleted(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.GetAdminScopeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.OrganizationStatusOffboarding), string(awstypes.OrganizationStatusOffboardingComplete)},
		Target:  []string{},
		Refresh: statusDelegatedAdministrator(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.GetAdminScopeOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDelegatedAdministrator(ctx context.Context, conn *fms.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDelegatedAdministratorByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findDelegatedAdministratorByID(ctx context.Context, conn *fms.Client, id string) (*fms.GetAdminScopeOutput, error) {
	in := &fms.GetAdminScopeInput{
		AdminAccount: aws.String(id),
	}

	out, err := conn.GetAdminScope(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AdminScope == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandAdminScope(ctx context.Context, in adminScope) *awstypes.AdminScope {
	out := &awstypes.AdminScope{
		AccountScope:            expandAccountScope(ctx, in.AccountScope),
		OrganizationalUnitScope: expandOrganizationalUnitScope(ctx, in.OrganizationalUnitScope),
		PolicyTypeScope:         expandPolicyTypeScope(ctx, in.PolicyTypeScope),
		RegionScope:             expandRegionScope(ctx, in.RegionScope),
	}
	return out
}

func expandAccountScope(ctx context.Context, in accountScope) *awstypes.AccountScope {
	out := &awstypes.AccountScope{
		Accounts:                 flex.ExpandFrameworkStringValueList(ctx, in.Accounts),
		AllAccountsEnabled:       in.AllAccountsEnabled.ValueBool(),
		ExcludeSpecifiedAccounts: in.ExcludeSpecifiedAccounts.ValueBool(),
	}
	return out
}

func expandOrganizationalUnitScope(ctx context.Context, in organizationalUnitScope) *awstypes.OrganizationalUnitScope {
	out := &awstypes.OrganizationalUnitScope{
		AllOrganizationalUnitsEnabled:       in.AllOrganizationalUnitsEnabled.ValueBool(),
		ExcludeSpecifiedOrganizationalUnits: in.ExcludeSpecifiedOrganizationalUnits.ValueBool(),
		OrganizationalUnits:                 flex.ExpandFrameworkStringValueList(ctx, in.OrganizationalUnits),
	}
	return out
}

func expandPolicyTypeScope(ctx context.Context, in policyTypeScope) *awstypes.PolicyTypeScope {
	var serviceTypes []awstypes.SecurityServiceType
	for _, t := range flex.ExpandFrameworkStringValueList(ctx, in.PolicyTypes) {
		serviceTypes = append(serviceTypes, awstypes.SecurityServiceType(t))
	}
	out := &awstypes.PolicyTypeScope{
		AllPolicyTypesEnabled: in.AllPolicyTypesEnabled.ValueBool(),
		PolicyTypes:           serviceTypes,
	}
	return out
}

func expandRegionScope(ctx context.Context, in regionScope) *awstypes.RegionScope {
	out := &awstypes.RegionScope{
		AllRegionsEnabled: in.AllRegionsEnabled.ValueBool(),
		Regions:           flex.ExpandFrameworkStringValueList(ctx, in.Regions),
	}
	return out
}

func flattenAdminScope(ctx context.Context, in *awstypes.AdminScope) adminScope {
	out := adminScope{
		AccountScope:            flattenAccountScope(ctx, in.AccountScope),
		OrganizationalUnitScope: flattenOrganizationalUnitScope(ctx, in.OrganizationalUnitScope),
		PolicyTypeScope:         flattenPolicyTypeScope(ctx, in.PolicyTypeScope),
		RegionScope:             flattenRegionScope(ctx, in.RegionScope),
	}
	return out
}

func flattenAccountScope(ctx context.Context, in *awstypes.AccountScope) accountScope {
	out := accountScope{
		Accounts:                 flex.FlattenFrameworkStringValueList(ctx, in.Accounts),
		AllAccountsEnabled:       types.BoolValue(in.AllAccountsEnabled),
		ExcludeSpecifiedAccounts: types.BoolValue(in.ExcludeSpecifiedAccounts),
	}
	return out
}

func flattenOrganizationalUnitScope(ctx context.Context, in *awstypes.OrganizationalUnitScope) organizationalUnitScope {
	out := organizationalUnitScope{
		AllOrganizationalUnitsEnabled:       types.BoolValue(in.AllOrganizationalUnitsEnabled),
		ExcludeSpecifiedOrganizationalUnits: types.BoolValue(in.ExcludeSpecifiedOrganizationalUnits),
		OrganizationalUnits:                 flex.FlattenFrameworkStringValueList(ctx, in.OrganizationalUnits),
	}
	return out

}

func flattenPolicyTypeScope(ctx context.Context, in *awstypes.PolicyTypeScope) policyTypeScope {
	out := policyTypeScope{
		AllPolicyTypesEnabled: types.BoolValue(in.AllPolicyTypesEnabled),
		PolicyTypes:           flex.FlattenFrameworkStringValueList(ctx, in.PolicyTypes),
	}
	return out
}

func flattenRegionScope(ctx context.Context, in *awstypes.RegionScope) regionScope {
	out := regionScope{
		AllRegionsEnabled: types.BoolValue(in.AllRegionsEnabled),
		Regions:           flex.FlattenFrameworkStringValueList(ctx, in.Regions),
	}
	return out
}

type resourceDelegatedAdministratorData struct {
	ID           types.String   `tfsdk:"id"`
	AdminAccount types.String   `tfsdk:"admin_account"`
	AdminScope   adminScope     `tfsdk:"admin_scope"`
	Status       types.String   `tfsdk:"status"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

type adminScope struct {
	AccountScope            accountScope            `tfsdk:"account_scope"`
	OrganizationalUnitScope organizationalUnitScope `tfsdk:"organizational_unit_scope"`
	PolicyTypeScope         policyTypeScope         `tfsdk:"policy_type_scope"`
	RegionScope             regionScope             `tfsdk:"region_scope"`
}

type accountScope struct {
	Accounts                 types.List
	AllAccountsEnabled       types.Bool
	ExcludeSpecifiedAccounts types.Bool
}

type organizationalUnitScope struct {
	AllOrganizationalUnitsEnabled       types.Bool
	ExcludeSpecifiedOrganizationalUnits types.Bool
	OrganizationalUnits                 types.List
}

type policyTypeScope struct {
	AllPolicyTypesEnabled types.Bool
	PolicyTypes           types.List
}

type regionScope struct {
	AllRegionsEnabled types.Bool
	Regions           types.List
}
