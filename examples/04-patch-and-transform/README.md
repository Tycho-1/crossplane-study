# Example 04: Patch & Transform (FromCompositeFieldPath, Combine)

Use **function-patch-and-transform** with multiple patches: **FromCompositeFieldPath** and (optional) **Combine** to build an S3 Bucket and set tags from the XR.

## Prerequisites

- Crossplane, provider-aws-s3, function-patch-and-transform installed
- ClusterProviderConfig with AWS credentials

## Apply

```bash
kubectl apply -f xrd.yaml
kubectl apply -f composition.yaml
kubectl apply -f xr.yaml
```

## Render locally

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

`functions.yaml` should declare the function (see below).

## functions.yaml (for crossplane render)

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.8.2
```

Save as `functions.yaml` in this directory, then:

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

## Cleanup

```bash
kubectl delete -f xr.yaml -f composition.yaml -f xrd.yaml
```
