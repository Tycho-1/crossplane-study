package main

import (
	"context"
	"testing"

	"dev.upbound.io/models/com/example/platform/v1alpha1"
	v1 "dev.upbound.io/models/io/k8s/meta/v1"
	"dev.upbound.io/models/io/upbound/aws/s3/v1beta1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composite"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"BucketNotYetCreated": {
			reason: "If the bucket hasn't been created yet, only the bucket should be desired.",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Observed: &fnv1.State{
						Composite: toResource(&v1alpha1.XStorageBucket{
							Spec: &v1alpha1.XStorageBucketSpec{
								Parameters: &v1alpha1.XStorageBucketSpecParameters{
									Region:     ptr.To("us-east-1"),
									ACL:        ptr.To("private"),
									Versioning: ptr.To(false),
								},
							},
						}),
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1.Result{{
						Severity: fnv1.Severity_SEVERITY_NORMAL,
						Message:  "waiting for bucket to be created",
						Target:   fnv1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
					}},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"bucket": toResource(&v1beta1.Bucket{
								APIVersion: ptr.To(v1beta1.BucketAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketKindBucket),
								Spec: &v1beta1.BucketSpec{
									ForProvider: &v1beta1.BucketSpecForProvider{
										Region: ptr.To("us-east-1"),
									},
								},
							}),
						},
					},
				},
			},
		},
		"BucketCreated": {
			reason: "If the bucket has been created, all resources should be desired.",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Observed: &fnv1.State{
						Composite: toResource(&v1alpha1.XStorageBucket{
							Spec: &v1alpha1.XStorageBucketSpec{
								Parameters: &v1alpha1.XStorageBucketSpecParameters{
									Region:     ptr.To("us-east-1"),
									ACL:        ptr.To("private"),
									Versioning: ptr.To(false),
								},
							},
						}),
						Resources: map[string]*fnv1.Resource{
							"bucket": toResource(&v1beta1.Bucket{
								APIVersion: ptr.To(v1beta1.BucketAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketKindBucket),
								Metadata: &v1.ObjectMeta{
									Annotations: &map[string]string{
										"crossplane.io/external-name": "my-bukkit",
									},
								},
								Spec: &v1beta1.BucketSpec{
									ForProvider: &v1beta1.BucketSpecForProvider{
										Region: ptr.To("us-east-1"),
									},
								},
							}),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta:    &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1.Result{},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"bucket": toResource(&v1beta1.Bucket{
								APIVersion: ptr.To(v1beta1.BucketAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketKindBucket),
								Spec: &v1beta1.BucketSpec{
									ForProvider: &v1beta1.BucketSpecForProvider{
										Region: ptr.To("us-east-1"),
									},
								},
							}),
							"acl": toResource(&v1beta1.BucketACL{
								APIVersion: ptr.To(v1beta1.BucketACLApiVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketACLKindBucketACL),
								Spec: &v1beta1.BucketACLSpec{
									ForProvider: &v1beta1.BucketACLSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										ACL:    ptr.To("private"),
									},
								},
							}),
							"boc": toResource(&v1beta1.BucketOwnershipControls{
								APIVersion: ptr.To(v1beta1.BucketOwnershipControlsAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketOwnershipControlsKindBucketOwnershipControls),
								Spec: &v1beta1.BucketOwnershipControlsSpec{
									ForProvider: &v1beta1.BucketOwnershipControlsSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										Rule: &[]v1beta1.BucketOwnershipControlsSpecForProviderRuleItem{{
											ObjectOwnership: ptr.To("BucketOwnerPreferred"),
										}},
									},
								},
							}),
							"pab": toResource(&v1beta1.BucketPublicAccessBlock{
								APIVersion: ptr.To(v1beta1.BucketPublicAccessBlockAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketPublicAccessBlockKindBucketPublicAccessBlock),
								Spec: &v1beta1.BucketPublicAccessBlockSpec{
									ForProvider: &v1beta1.BucketPublicAccessBlockSpecForProvider{
										Bucket:                ptr.To("my-bukkit"),
										Region:                ptr.To("us-east-1"),
										BlockPublicAcls:       ptr.To(false),
										RestrictPublicBuckets: ptr.To(false),
										IgnorePublicAcls:      ptr.To(false),
										BlockPublicPolicy:     ptr.To(false),
									},
								},
							}),
							"sse": toResource(&v1beta1.BucketServerSideEncryptionConfiguration{
								APIVersion: ptr.To(v1beta1.BucketServerSideEncryptionConfigurationAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketServerSideEncryptionConfigurationKindBucketServerSideEncryptionConfiguration),
								Spec: &v1beta1.BucketServerSideEncryptionConfigurationSpec{
									ForProvider: &v1beta1.BucketServerSideEncryptionConfigurationSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										Rule: &[]v1beta1.BucketServerSideEncryptionConfigurationSpecForProviderRuleItem{{
											ApplyServerSideEncryptionByDefault: &[]v1beta1.BucketServerSideEncryptionConfigurationSpecForProviderRuleItemApplyServerSideEncryptionByDefaultItem{{
												SseAlgorithm: ptr.To("AES256"),
											}},
											BucketKeyEnabled: ptr.To(true),
										}},
									},
								},
							}),
						},
					},
				},
			},
		},
		"BucketCreatedWithVersioning": {
			reason: "If the bucket has been created with versioning, all resources should be desired.",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Observed: &fnv1.State{
						Composite: toResource(&v1alpha1.XStorageBucket{
							Spec: &v1alpha1.XStorageBucketSpec{
								Parameters: &v1alpha1.XStorageBucketSpecParameters{
									Region:     ptr.To("us-east-1"),
									ACL:        ptr.To("private"),
									Versioning: ptr.To(true),
								},
							},
						}),
						Resources: map[string]*fnv1.Resource{
							"bucket": toResource(&v1beta1.Bucket{
								APIVersion: ptr.To(v1beta1.BucketAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketKindBucket),
								Metadata: &v1.ObjectMeta{
									Annotations: &map[string]string{
										"crossplane.io/external-name": "my-bukkit",
									},
								},
								Spec: &v1beta1.BucketSpec{
									ForProvider: &v1beta1.BucketSpecForProvider{
										Region: ptr.To("us-east-1"),
									},
								},
							}),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta:    &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1.Result{},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"bucket": toResource(&v1beta1.Bucket{
								APIVersion: ptr.To(v1beta1.BucketAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketKindBucket),
								Spec: &v1beta1.BucketSpec{
									ForProvider: &v1beta1.BucketSpecForProvider{
										Region: ptr.To("us-east-1"),
									},
								},
							}),
							"acl": toResource(&v1beta1.BucketACL{
								APIVersion: ptr.To(v1beta1.BucketACLApiVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketACLKindBucketACL),
								Spec: &v1beta1.BucketACLSpec{
									ForProvider: &v1beta1.BucketACLSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										ACL:    ptr.To("private"),
									},
								},
							}),
							"boc": toResource(&v1beta1.BucketOwnershipControls{
								APIVersion: ptr.To(v1beta1.BucketOwnershipControlsAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketOwnershipControlsKindBucketOwnershipControls),
								Spec: &v1beta1.BucketOwnershipControlsSpec{
									ForProvider: &v1beta1.BucketOwnershipControlsSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										Rule: &[]v1beta1.BucketOwnershipControlsSpecForProviderRuleItem{{
											ObjectOwnership: ptr.To("BucketOwnerPreferred"),
										}},
									},
								},
							}),
							"pab": toResource(&v1beta1.BucketPublicAccessBlock{
								APIVersion: ptr.To(v1beta1.BucketPublicAccessBlockAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketPublicAccessBlockKindBucketPublicAccessBlock),
								Spec: &v1beta1.BucketPublicAccessBlockSpec{
									ForProvider: &v1beta1.BucketPublicAccessBlockSpecForProvider{
										Bucket:                ptr.To("my-bukkit"),
										Region:                ptr.To("us-east-1"),
										BlockPublicAcls:       ptr.To(false),
										RestrictPublicBuckets: ptr.To(false),
										IgnorePublicAcls:      ptr.To(false),
										BlockPublicPolicy:     ptr.To(false),
									},
								},
							}),
							"sse": toResource(&v1beta1.BucketServerSideEncryptionConfiguration{
								APIVersion: ptr.To(v1beta1.BucketServerSideEncryptionConfigurationAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketServerSideEncryptionConfigurationKindBucketServerSideEncryptionConfiguration),
								Spec: &v1beta1.BucketServerSideEncryptionConfigurationSpec{
									ForProvider: &v1beta1.BucketServerSideEncryptionConfigurationSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										Rule: &[]v1beta1.BucketServerSideEncryptionConfigurationSpecForProviderRuleItem{{
											ApplyServerSideEncryptionByDefault: &[]v1beta1.BucketServerSideEncryptionConfigurationSpecForProviderRuleItemApplyServerSideEncryptionByDefaultItem{{
												SseAlgorithm: ptr.To("AES256"),
											}},
											BucketKeyEnabled: ptr.To(true),
										}},
									},
								},
							}),
							"versioning": toResource(&v1beta1.BucketVersioning{
								APIVersion: ptr.To(v1beta1.BucketVersioningAPIVersions3AwsUpboundIoV1Beta1),
								Kind:       ptr.To(v1beta1.BucketVersioningKindBucketVersioning),
								Spec: &v1beta1.BucketVersioningSpec{
									ForProvider: &v1beta1.BucketVersioningSpecForProvider{
										Bucket: ptr.To("my-bukkit"),
										Region: ptr.To("us-east-1"),
										VersioningConfiguration: &[]v1beta1.BucketVersioningSpecForProviderVersioningConfigurationItem{{
											Status: ptr.To("Enabled"),
										}},
									},
								},
							}),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}

func toResource(in any) *fnv1.Resource {
	obj := composite.New()
	_ = convertViaJSON(obj, in)
	pb, _ := resource.AsStruct(obj)
	return &fnv1.Resource{
		Resource: pb,
	}
}
