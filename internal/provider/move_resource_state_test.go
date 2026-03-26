// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// rawStateFromMap builds a *tfprotov5.RawState from a map of string values.
// All values are JSON-encoded; nil means JSON null.
func rawStateFromMap(t *testing.T, m map[string]any) *tfprotov5.RawState {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal: %s", err)
	}
	return &tfprotov5.RawState{JSON: b}
}

func TestMoveS3BucketObjectToS3ObjectState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	srcRaw := rawStateFromMap(t, map[string]any{
		"id":                            "my-bucket/my-key",
		"acl":                           "private",
		"arn":                           "arn:aws:s3:::my-bucket/my-key",
		"bucket":                        "my-bucket",
		"bucket_key_enabled":            false,
		"cache_control":                 nil,
		"content":                       "hello world",
		"content_base64":                nil,
		"content_disposition":           nil,
		"content_encoding":              nil,
		"content_language":              nil,
		"content_type":                  "text/plain",
		"etag":                          "abc123",
		"force_destroy":                 false,
		"key":                           "my-key",
		"kms_key_id":                    nil,
		"metadata":                      map[string]any{},
		"object_lock_legal_hold_status": nil,
		"object_lock_mode":              nil,
		"object_lock_retain_until_date": nil,
		"server_side_encryption":        nil,
		"source":                        nil,
		"source_hash":                   nil,
		"storage_class":                 "STANDARD",
		"tags":                          map[string]any{},
		"tags_all":                      map[string]any{},
		"version_id":                    nil,
		"website_redirect":              nil,
	})

	req := &tfprotov5.MoveResourceStateRequest{
		SourceProviderAddress: "registry.terraform.io/hashicorp/aws",
		SourceTypeName:        "aws_s3_bucket_object",
		SourceSchemaVersion:   0,
		TargetTypeName:        "aws_s3_object",
		SourceState:           srcRaw,
	}

	resp, err := moveS3BucketObjectToS3ObjectState(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(resp.Diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.TargetState == nil {
		t.Fatal("expected TargetState to be set")
	}

	// Decode the target state and verify fields.
	targetVal, err := resp.TargetState.Unmarshal(s3ObjectStateType)
	if err != nil {
		t.Fatalf("decoding target state: %s", err)
	}

	var targetAttrs map[string]tftypes.Value
	if err := targetVal.As(&targetAttrs); err != nil {
		t.Fatalf("extracting target attributes: %s", err)
	}

	// Verify copied fields.
	checkStringAttr(t, targetAttrs, "id", "my-bucket/my-key")
	checkStringAttr(t, targetAttrs, "bucket", "my-bucket")
	checkStringAttr(t, targetAttrs, "key", "my-key")
	checkStringAttr(t, targetAttrs, "acl", "private")
	checkStringAttr(t, targetAttrs, "arn", "arn:aws:s3:::my-bucket/my-key")
	checkStringAttr(t, targetAttrs, "content", "hello world")
	checkStringAttr(t, targetAttrs, "content_type", "text/plain")
	checkStringAttr(t, targetAttrs, "etag", "abc123")
	checkStringAttr(t, targetAttrs, "storage_class", "STANDARD")

	// Verify new fields are null (checksum_*).
	for _, k := range []string{
		"checksum_algorithm",
		"checksum_crc32",
		"checksum_crc32c",
		"checksum_crc64nvme",
		"checksum_sha1",
		"checksum_sha256",
	} {
		v, ok := targetAttrs[k]
		if !ok {
			t.Errorf("expected %q in target attributes", k)
			continue
		}
		if !v.IsNull() {
			t.Errorf("expected %q to be null, got %s", k, v)
		}
	}

	// Verify override_provider is an empty list.
	v, ok := targetAttrs["override_provider"]
	if !ok {
		t.Fatal("expected \"override_provider\" in target attributes")
	}
	var overrideList []tftypes.Value
	if err := v.As(&overrideList); err != nil {
		t.Fatalf("decoding override_provider: %s", err)
	}
	if len(overrideList) != 0 {
		t.Errorf("expected override_provider to be empty list, got %d elements", len(overrideList))
	}
}

func TestMoveS3BucketObjectToS3ObjectState_WrongSchemaVersion(t *testing.T) {
	t.Parallel()

	req := &tfprotov5.MoveResourceStateRequest{
		SourceProviderAddress: "registry.terraform.io/hashicorp/aws",
		SourceTypeName:        "aws_s3_bucket_object",
		SourceSchemaVersion:   1, // unexpected version
		TargetTypeName:        "aws_s3_object",
		SourceState:           &tfprotov5.RawState{JSON: []byte(`{}`)},
	}

	resp, err := moveS3BucketObjectToS3ObjectState(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected diagnostics for unexpected schema version")
	}
}

func TestMoveS3BucketObjectToS3ObjectState_WrongProvider(t *testing.T) {
	t.Parallel()

	req := &tfprotov5.MoveResourceStateRequest{
		SourceProviderAddress: "registry.terraform.io/other/aws",
		SourceTypeName:        "aws_s3_bucket_object",
		SourceSchemaVersion:   0,
		TargetTypeName:        "aws_s3_object",
		SourceState:           &tfprotov5.RawState{JSON: []byte(`{}`)},
	}

	resp, err := moveS3BucketObjectToS3ObjectState(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected diagnostics for wrong provider address")
	}
}

func TestMoveResourceStateServer_DelegatesToUnderlying(t *testing.T) {
	t.Parallel()

	// Ensure the wrapper delegates non-S3 migration calls to the underlying server.
	underlying := &mockProviderServer{}
	s := &moveResourceStateServer{ProviderServer: underlying}

	req := &tfprotov5.MoveResourceStateRequest{
		TargetTypeName: "aws_other_resource",
		SourceTypeName: "aws_other_resource_old",
	}

	_, _ = s.MoveResourceState(context.Background(), req)

	if !underlying.moveResourceStateCalled {
		t.Error("expected underlying MoveResourceState to be called for non-S3 migration")
	}
}

// mockProviderServer is a minimal mock that records MoveResourceState calls.
type mockProviderServer struct {
	tfprotov5.ProviderServer
	moveResourceStateCalled bool
}

func (m *mockProviderServer) MoveResourceState(_ context.Context, _ *tfprotov5.MoveResourceStateRequest) (*tfprotov5.MoveResourceStateResponse, error) {
	m.moveResourceStateCalled = true
	return &tfprotov5.MoveResourceStateResponse{}, nil
}

// checkStringAttr asserts that attrs[name] is a non-null string equal to want.
func checkStringAttr(t *testing.T, attrs map[string]tftypes.Value, name, want string) {
	t.Helper()
	v, ok := attrs[name]
	if !ok {
		t.Errorf("expected attribute %q", name)
		return
	}
	var got string
	if err := v.As(&got); err != nil {
		t.Errorf("attribute %q: %s", name, err)
		return
	}
	if got != want {
		t.Errorf("attribute %q: got %q, want %q", name, got, want)
	}
}
