# Example 01: Simple S3 Bucket (Managed Resource)

Create a single S3 bucket using the **provider-aws-s3** managed resource (v2 namespaced style).

## Prerequisites

- Crossplane installed
- **provider-aws-s3** installed and **HEALTHY**
- **ClusterProviderConfig** or **ProviderConfig** with valid AWS credentials (see [02-provider-config-aws](../02-provider-config-aws/))

## Apply

```bash
kubectl apply -f bucket.yaml
```

## Verify

```bash
kubectl get bucket -n default
kubectl describe bucket crossplane-bucket-demo -n default
```

Check AWS:

```bash
aws s3 ls | grep crossplane-bucket-demo
```

## Cleanup

```bash
kubectl delete -f bucket.yaml
```

The provider will delete the bucket in AWS when the MR is removed.
