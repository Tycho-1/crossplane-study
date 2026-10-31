# my-platform Configuration

A Crossplane **Configuration** package that provides the `XPlatform` API.  
One XR creates three AWS resources: an S3 Bucket, a DynamoDB Table, and an IAM Policy.

## Install

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
spec:
  package: docker.io/yourorg/my-platform:v0.1.0
EOF

# Watch until INSTALLED=True and HEALTHY=True
kubectl get configurations.pkg.crossplane.io --watch
```

## Prerequisites

The Configuration automatically installs its dependencies via `dependsOn`. You only need
to provide a `ProviderConfig` with valid AWS credentials before creating XR instances:

```bash
kubectl apply -f provider-config.yaml   # your AWS credentials secret + ProviderConfig
```

## Parameters

All fields live under `spec` of the `XPlatform` XR.

| Field | Type | Default | Required | Description |
|---|---|---|---|---|
| `spec.region` | string | `us-east-2` | no | AWS region for all resources |
| `spec.storageName` | string | — | **yes** | Name of the S3 Bucket |
| `spec.tableName` | string | — | **yes** | Name of the DynamoDB Table |
| `spec.iamPolicyName` | string | — | **yes** | Name of the IAM Policy |

## What gets deployed

One `XPlatform` XR creates exactly three AWS managed resources:

| Resource | AWS service | Notes |
|---|---|---|
| S3 Bucket | `s3.aws.m.upbound.io/v1beta1 Bucket` | Region patched from `spec.region`, name from `spec.storageName` |
| DynamoDB Table | `dynamodb.aws.m.upbound.io/v1beta1 Table` | `PAY_PER_REQUEST` billing, `LockID` (String) hash key — suitable as a Terraform state-lock table |
| IAM Policy | `iam.aws.m.upbound.io/v1beta1 Policy` | Grants `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject` on `arn:aws:s3:::*` |

## Usage example

```yaml
apiVersion: example.org/v1alpha1
kind: XPlatform
metadata:
  namespace: default
  name: my-platform
spec:
  region: eu-central-1
  storageName: my-platform-storage
  tableName: my-platform-lock-table
  iamPolicyName: my-platform-s3-policy
```

```bash
# Same XR as above (kept outside the package root so xpkg build ignores it)
kubectl apply -f ../examples/07-multi-resource-composition/xr.yaml

# Verify
kubectl get xplatforms -n default
kubectl get bucket
kubectl get table.dynamodb
kubectl get policy.iam
```

## Dry-run locally (no cluster needed)

```bash
crossplane render \
  ../examples/07-multi-resource-composition/xr.yaml \
  composition.yaml \
  ../examples/07-multi-resource-composition/functions.yaml
```

## Build and push

```bash
crossplane xpkg build --package-root=. --output=my-platform.xpkg
crossplane xpkg push --package=my-platform.xpkg docker.io/yourorg/my-platform:v0.1.0
```

## Package contents

| File | Purpose |
|---|---|
| `crossplane.yaml` | Package metadata and `dependsOn` (Providers + Function auto-installed) |
| `xrd.yaml` | `XPlatform` API definition (schema, required fields) |
| `composition.yaml` | Pipeline: `function-patch-and-transform` creates the 3 resources |

## Dependencies (auto-installed)

| Package | Version | Purpose |
|---|---|---|
| `function-patch-and-transform` | `>=v0.10.0` | Composition pipeline function |
| `provider-aws-s3` | `>=v2.0.0` | S3 Bucket managed resource |
| `provider-aws-dynamodb` | `>=v2.0.0` | DynamoDB Table managed resource |
| `provider-aws-iam` | `>=v2.0.0` | IAM Policy managed resource |

## Cleanup

```bash
kubectl delete xplatform my-platform -n default
# Wait for AWS resources to be deleted, then:
kubectl delete -f configuration-install.yaml
```
