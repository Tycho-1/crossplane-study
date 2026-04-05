package crossplane

// XPlatform Composition — uses params from values.cue
#Composition: {
	apiVersion: "apiextensions.crossplane.io/v1"
	kind:       "Composition"
	metadata: {
		name: params.compositionName
	}
	spec: {
		compositeTypeRef: {
			apiVersion: "\(params.group)/v1alpha1"
			kind:       params.kind
		}
		mode: "Pipeline"
		pipeline: [{
			step: "create-resources"
			functionRef: {
				name: "function-patch-and-transform"
			}
			input: {
				apiVersion: "pt.fn.crossplane.io/v1beta1"
				kind:       "Resources"
				resources: [
					{
						name: "bucket"
						base: {
							apiVersion: "s3.aws.m.upbound.io/v1beta1"
							kind:       "Bucket"
							spec: {
								forProvider: {
									region: params.defaultRegion
								}
							}
						}
						patches: [
							{
								type:         "FromCompositeFieldPath"
								fromFieldPath: "spec.region"
								toFieldPath:   "spec.forProvider.region"
							},
							{
								type:         "FromCompositeFieldPath"
								fromFieldPath: "spec.storageName"
								toFieldPath:   "metadata.name"
								policy: {
									fromFieldPath: "Required"
								}
							},
						]
					},
					{
						name: "dynamodb-table"
						base: {
							apiVersion: "dynamodb.aws.m.upbound.io/v1beta1"
							kind:       "Table"
							spec: {
								forProvider: {
									region:      params.defaultRegion
									billingMode: "PAY_PER_REQUEST"
									hashKey:     "LockID"
									attribute: [{
										name: "LockID"
										type: "S"
									}]
								}
							}
						}
						patches: [
							{
								type:         "FromCompositeFieldPath"
								fromFieldPath: "spec.region"
								toFieldPath:   "spec.forProvider.region"
							},
							{
								type:         "FromCompositeFieldPath"
								fromFieldPath: "spec.tableName"
								toFieldPath:   "metadata.name"
								policy: {
									fromFieldPath: "Required"
								}
							},
						]
					},
					{
						name: "iam-policy"
						base: {
							apiVersion: "iam.aws.m.upbound.io/v1beta1"
							kind:       "Policy"
							spec: {
								forProvider: {
									description: "S3 read/write access"
									policy: """
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:PutObject","s3:GetObject","s3:DeleteObject"],
    "Resource": "arn:aws:s3:::*"
  }]
}
"""
								}
							}
						}
						patches: [{
							type:         "FromCompositeFieldPath"
							fromFieldPath: "spec.iamPolicyName"
							toFieldPath:   "metadata.name"
							policy: {
								fromFieldPath: "Required"
							}
						}]
					},
				]
			}
		}]
	}
}

composition: #Composition
