# Example 08: Kubernetes application (v2)

**Application example:** Crossplane is not only for cloud resources (S3, RDS). In v2 you can use it for **normal Kubernetes applications**. This example defines an **XApp** that creates a **Deployment** and a **Service**—a simple application in the cluster with no cloud MRs.

Same idea as [05-app-with-deployment](../05-app-with-deployment/); this folder is the dedicated **application** example.

## Prerequisites

- Crossplane installed (v2)
- **function-patch-and-transform** installed
- RBAC: Crossplane must be able to create `deployments` and `services`. If not, apply `rbac.yaml`.

## RBAC (if needed)

```bash
kubectl apply -f rbac.yaml
```

## Apply

```bash
kubectl apply -f xrd.yaml
kubectl apply -f composition.yaml
kubectl apply -f xr.yaml
```

## Verify

```bash
kubectl get xapp -n default
kubectl get deployment,service -n default -l crossplane.io/composite=hello-app
kubectl get pods -n default -l app=hello-app
```

## Cleanup

```bash
kubectl delete -f xr.yaml
kubectl delete -f composition.yaml
kubectl delete -f xrd.yaml
# optionally: kubectl delete -f rbac.yaml
```
