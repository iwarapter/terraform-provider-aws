// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// moveResourceStateServer wraps a ProviderServer to implement MoveResourceState
// for specific cross-type resource migrations that are not supported by the
// underlying servers (e.g. terraform-plugin-sdk/v2 always returns "not supported").
type moveResourceStateServer struct {
	tfprotov5.ProviderServer
}

// MoveResourceState handles specific cross-type resource state migrations.
// For all other cases it delegates to the underlying provider server.
func (s *moveResourceStateServer) MoveResourceState(ctx context.Context, req *tfprotov5.MoveResourceStateRequest) (*tfprotov5.MoveResourceStateResponse, error) {
	if req == nil {
		return s.ProviderServer.MoveResourceState(ctx, req)
	}

	if req.TargetTypeName == "aws_s3_object" && req.SourceTypeName == "aws_s3_bucket_object" {
		return moveS3BucketObjectToS3ObjectState(ctx, req)
	}

	return s.ProviderServer.MoveResourceState(ctx, req)
}

// s3BucketObjectStateType is the tftypes.Object schema for aws_s3_bucket_object state.
// It must exactly match what terraform-plugin-sdk/v2 encodes for that resource.
var s3BucketObjectStateType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":                            tftypes.String,
		"acl":                           tftypes.String,
		"arn":                           tftypes.String,
		"bucket":                        tftypes.String,
		"bucket_key_enabled":            tftypes.Bool,
		"cache_control":                 tftypes.String,
		"content":                       tftypes.String,
		"content_base64":                tftypes.String,
		"content_disposition":           tftypes.String,
		"content_encoding":              tftypes.String,
		"content_language":              tftypes.String,
		"content_type":                  tftypes.String,
		"etag":                          tftypes.String,
		"force_destroy":                 tftypes.Bool,
		"key":                           tftypes.String,
		"kms_key_id":                    tftypes.String,
		"metadata":                      tftypes.Map{ElementType: tftypes.String},
		"object_lock_legal_hold_status": tftypes.String,
		"object_lock_mode":              tftypes.String,
		"object_lock_retain_until_date": tftypes.String,
		"region":                        tftypes.String,
		"server_side_encryption":        tftypes.String,
		"source":                        tftypes.String,
		"source_hash":                   tftypes.String,
		"storage_class":                 tftypes.String,
		"tags":                          tftypes.Map{ElementType: tftypes.String},
		"tags_all":                      tftypes.Map{ElementType: tftypes.String},
		"version_id":                    tftypes.String,
		"website_redirect":              tftypes.String,
	},
}

// s3ObjectOverrideProviderType is the tftypes for the override_provider list element.
var s3ObjectOverrideProviderType = tftypes.List{
	ElementType: tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"default_tags": tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.Map{ElementType: tftypes.String},
					},
				},
			},
		},
	},
}

// s3ObjectStateType is the tftypes.Object schema for aws_s3_object state.
// It must exactly match what terraform-plugin-sdk/v2 encodes for that resource.
var s3ObjectStateType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":                            tftypes.String,
		"acl":                           tftypes.String,
		"arn":                           tftypes.String,
		"bucket":                        tftypes.String,
		"bucket_key_enabled":            tftypes.Bool,
		"cache_control":                 tftypes.String,
		"checksum_algorithm":            tftypes.String,
		"checksum_crc32":                tftypes.String,
		"checksum_crc32c":               tftypes.String,
		"checksum_crc64nvme":            tftypes.String,
		"checksum_sha1":                 tftypes.String,
		"checksum_sha256":               tftypes.String,
		"content":                       tftypes.String,
		"content_base64":                tftypes.String,
		"content_disposition":           tftypes.String,
		"content_encoding":              tftypes.String,
		"content_language":              tftypes.String,
		"content_type":                  tftypes.String,
		"etag":                          tftypes.String,
		"force_destroy":                 tftypes.Bool,
		"key":                           tftypes.String,
		"kms_key_id":                    tftypes.String,
		"metadata":                      tftypes.Map{ElementType: tftypes.String},
		"object_lock_legal_hold_status": tftypes.String,
		"object_lock_mode":              tftypes.String,
		"object_lock_retain_until_date": tftypes.String,
		"override_provider":             s3ObjectOverrideProviderType,
		"region":                        tftypes.String,
		"server_side_encryption":        tftypes.String,
		"source":                        tftypes.String,
		"source_hash":                   tftypes.String,
		"storage_class":                 tftypes.String,
		"tags":                          tftypes.Map{ElementType: tftypes.String},
		"tags_all":                      tftypes.Map{ElementType: tftypes.String},
		"version_id":                    tftypes.String,
		"website_redirect":              tftypes.String,
	},
}

// moveS3BucketObjectToS3ObjectState translates terraform state from
// aws_s3_bucket_object to aws_s3_object.
func moveS3BucketObjectToS3ObjectState(_ context.Context, req *tfprotov5.MoveResourceStateRequest) (*tfprotov5.MoveResourceStateResponse, error) {
	resp := &tfprotov5.MoveResourceStateResponse{}

	if !strings.HasSuffix(req.SourceProviderAddress, "hashicorp/aws") {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unexpected source provider",
			Detail:   fmt.Sprintf("Expected source provider address to end with \"hashicorp/aws\", got: %s", req.SourceProviderAddress),
		}}
		return resp, nil
	}

	if req.SourceSchemaVersion != 0 {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Unexpected source schema version",
			Detail:   fmt.Sprintf("Expected aws_s3_bucket_object schema version 0, got: %d", req.SourceSchemaVersion),
		}}
		return resp, nil
	}

	if req.SourceState == nil {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Missing source state",
			Detail:   "The source state for aws_s3_bucket_object is nil.",
		}}
		return resp, nil
	}

	// Decode the source (aws_s3_bucket_object) state.
	srcVal, err := req.SourceState.Unmarshal(s3BucketObjectStateType)
	if err != nil {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to decode source state",
			Detail:   fmt.Sprintf("Error decoding aws_s3_bucket_object state: %s", err),
		}}
		return resp, nil
	}

	var srcAttrs map[string]tftypes.Value
	if err := srcVal.As(&srcAttrs); err != nil {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to extract source attributes",
			Detail:   fmt.Sprintf("Error extracting aws_s3_bucket_object attributes: %s", err),
		}}
		return resp, nil
	}

	// Build the target (aws_s3_object) attributes by copying matching fields.
	targetAttrs := make(map[string]tftypes.Value, len(s3ObjectStateType.AttributeTypes))

	for k := range s3ObjectStateType.AttributeTypes {
		if v, ok := srcAttrs[k]; ok {
			targetAttrs[k] = v
		}
	}

	// Set new aws_s3_object-only fields that did not exist in aws_s3_bucket_object.
	// Checksum fields are Computed; they will be populated on the next refresh.
	targetAttrs["checksum_algorithm"] = tftypes.NewValue(tftypes.String, nil)
	targetAttrs["checksum_crc32"] = tftypes.NewValue(tftypes.String, nil)
	targetAttrs["checksum_crc32c"] = tftypes.NewValue(tftypes.String, nil)
	targetAttrs["checksum_crc64nvme"] = tftypes.NewValue(tftypes.String, nil)
	targetAttrs["checksum_sha1"] = tftypes.NewValue(tftypes.String, nil)
	targetAttrs["checksum_sha256"] = tftypes.NewValue(tftypes.String, nil)
	// override_provider is Optional; set to empty list (unconfigured).
	targetAttrs["override_provider"] = tftypes.NewValue(s3ObjectOverrideProviderType, []tftypes.Value{})

	targetVal := tftypes.NewValue(s3ObjectStateType, targetAttrs)

	targetState, err := tfprotov5.NewDynamicValue(s3ObjectStateType, targetVal)
	if err != nil {
		resp.Diagnostics = []*tfprotov5.Diagnostic{{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Failed to encode target state",
			Detail:   fmt.Sprintf("Error encoding aws_s3_object state: %s", err),
		}}
		return resp, nil
	}

	resp.TargetState = &targetState
	return resp, nil
}
