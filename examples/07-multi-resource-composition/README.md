# Example 07: Multi-resource Composition (Bucket + DynamoDB + IAM Policy)

One XR (`XPlatform`) that composes **three** managed resources in a single Composition:

| Resource | Provider CRD | Purpose |
|---|---|---|
| S3 Bucket | `s3.aws.m.upbound.io/v1beta1 Bucket` | Object storage |
| DynamoDB Table | `dynamodb.aws.m.upbound.io/v1beta1 Table` | State-lock table (PAY_PER_REQUEST, `LockID` hash key) |
| IAM Policy | `iam.aws.m.upbound.io/v1beta1 Policy` | S3 read/write permissions for the bucket |

## Prerequisites

- Crossplane installed in the cluster
- Providers: `provider-aws-s3`, `provider-aws-dynamodb`, `provider-aws-iam`
- `function-patch-and-transform` installed
- A `ClusterProviderConfig` with valid AWS credentials

## XR fields

| Field | Description |
|---|---|
| `spec.region` | AWS region for Bucket and DynamoDB table |
| `spec.storageName` | Name of the S3 Bucket |
| `spec.tableName` | Name of the DynamoDB Table |
| `spec.iamPolicyName` | Name of the IAM Policy |

## Dry-run locally (no cluster needed)

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

Crossplane pulls `function-patch-and-transform` via Docker and prints the composed resources to stdout.

> **Note: `xrd.yaml` is NOT required for `crossplane render`.**
> The render command only runs the Composition pipeline and shows what managed resources would be produced.
> It does not validate the XR against the schema — it simply trusts the fields in `xr.yaml`.
>
> The XRD is only needed when applying to a real cluster, where Crossplane uses it to:
> - validate that XR fields match the declared schema
> - register the custom resource so `kubectl apply -f xr.yaml` is accepted by the API server
>
> | Command | `xrd.yaml` needed? |
> |---|---|
> | `crossplane render` (local dry-run) | No |
> | `kubectl apply` (real cluster) | Yes |

## Apply

```bash
kubectl apply -f xrd.yaml
kubectl apply -f composition.yaml
kubectl apply -f xr.yaml
```

## Verify

```bash
kubectl get xplatform -n default
kubectl get bucket
kubectl get table.dynamodb
kubectl get policy.iam
kubectl describe xplatform my-platform -n default
```

## Cleanup

```bash
kubectl delete -f xr.yaml -f composition.yaml -f xrd.yaml
```

---

## How to add another resource

Adding a fourth (or fifth, ...) resource requires a change in each of the three files.

### Step 1 — `xrd.yaml`: declare the new field

Add a new property under `spec.versions[].schema.openAPIV3Schema.properties.spec.properties`
and optionally add it to the `required` list:

```yaml
properties:
  myNewResourceName:       # new field
    type: string
required:
  - myNewResourceName      # optional: make it required
```

### Step 2 — `composition.yaml`: add the new resource block

Append a new entry to the `resources` list inside the pipeline step.
Copy the pattern from an existing resource and adjust `apiVersion`, `kind`, and `fromFieldPath`:

```yaml
resources:
  # ... existing resources ...

  - name: my-new-resource                         # logical name used by Crossplane internally
    base:
      apiVersion: <group>/<version>               # CRD apiVersion from the provider
      kind: <Kind>                                # CRD kind from the provider
      spec:
        forProvider:
          region: us-east-2                       # static defaults
    patches:
      - type: FromCompositeFieldPath
        fromFieldPath: spec.myNewResourceName     # field declared in xrd.yaml (Step 1)
        toFieldPath: metadata.name
        policy:
          fromFieldPath: Required
```

### Step 3 — `xr.yaml`: provide the value

Add the field value to your XR instance:

```yaml
spec:
  myNewResourceName: my-platform-new-resource
```

That's it — Crossplane will create the new managed resource alongside the existing ones whenever the XR is reconciled.
