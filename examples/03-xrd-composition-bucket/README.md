# Example 03: Custom API (XRD) + Composition → S3 Bucket

Define a **custom API** (`XBucket`) and a **Composition** that creates an S3 Bucket when a user creates an XR. Uses **function-patch-and-transform**.

## Prerequisites

- Crossplane installed
- **provider-aws-s3** installed and healthy
- **function-patch-and-transform** installed (`kubectl apply -f functions.yaml`)
- **ClusterProviderConfig** with AWS credentials (see [02-provider-config-aws](../02-provider-config-aws/))

## Apply order

1. XRD (defines the API)
2. Composition (implements it with a function pipeline)
3. XR (instance of the API)

```bash
kubectl apply -f xrd.yaml
kubectl apply -f composition.yaml
kubectl apply -f xr.yaml
```

Or apply the directory (order may matter; apply XRD first):

```bash
kubectl apply -f xrd.yaml -f composition.yaml -f xr.yaml
```

## Verify

```bash
kubectl get xbucket
kubectl get bucket -n default
kubectl describe xbucket my-bucket -n default
```

## Render locally (no cluster)

If you have the Crossplane CLI and Docker:

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

[functions.yaml](./functions.yaml) declares the function-patch-and-transform Function so the CLI knows which image to run.

## Cleanup

```bash
kubectl delete -f xr.yaml
kubectl delete -f composition.yaml
kubectl delete -f xrd.yaml
```
