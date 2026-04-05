# Example 05: v2 – Compose Deployment + Service + ConfigMap (any K8s resource)

**v2 feature:** A Composition can compose **any** Kubernetes resource, not only Crossplane MRs. This example defines an **XApp** that creates a **Deployment**, a **Service**, and a **ConfigMap** in the same namespace via a **function pipeline** (Patch & Transform).

You must **grant Crossplane RBAC** to create Deployments, Services, and ConfigMaps (see below). Core resources are often already allowed; if not, apply `rbac.yaml`.

## Concepts (XRD, Composition, XR, CRD, Function)

For a detailed explanation of **what an XRD is**, how it **relates to Composition and XR**, **how a CRD is created**, and **how the function works** in this example, see **[EXPLANATION.md](./EXPLANATION.md)**.

- **XRD** – Defines the custom API (XApp with `spec.image`, `spec.replicas`) and causes Crossplane to create a **CRD** for it.
- **Composition** – Defines how that API is implemented: a **pipeline** of steps, each calling a **Function** (e.g. Patch & Transform) with input (base resources + patches).
- **XR** – An instance of the API (e.g. `XApp` named `my-app`). Creating/updating the XR triggers the pipeline; the function outputs Deployment, Service, ConfigMap; Crossplane applies them.
- **Function** – A step in the pipeline (e.g. `function-patch-and-transform`). The Composition references it by name; it must be installed via a **Function** resource (`functions.yaml`).

## Prerequisites

- Crossplane installed (v2)
- RBAC: Crossplane must be able to create `deployments`, `services`, and `configmaps` in the target namespace. If your install doesn’t grant this by default, apply `rbac.yaml`.

## RBAC (if needed)

If Crossplane doesn’t already have access to Deployments/Services/ConfigMaps:

```bash
kubectl apply -f rbac.yaml
```

## Apply order

1. **XRD** – creates the CRD for XApp.
2. **Function** – installs the Patch & Transform function so the pipeline can call it.
3. **Composition** – registers the pipeline for XApp (two steps: Deployment+Service, then ConfigMap).
4. **XR** – creates an XApp instance; Crossplane runs the pipeline and creates the composed resources.

```bash
kubectl apply -f xrd.yaml
kubectl apply -f functions.yaml
kubectl get functions   # wait until function-patch-and-transform is INSTALLED and HEALTHY
kubectl apply -f composition.yaml
kubectl apply -f xr.yaml
```

## Verify

```bash
kubectl get xapp -n default
kubectl get deployment,service,configmap -n default -l crossplane.io/composite=my-app
kubectl describe xapp my-app -n default
```

## Why READY is False (and how to fix it)

If the XR shows **SYNCED: True** but **READY: False** while the Deployment, Service, and ConfigMap exist and look fine, the reason is:

**Crossplane does not infer readiness for plain Kubernetes resources** (Deployment, Service, ConfigMap). By default it only marks the XR ready when composed resources report a “Ready” condition; plain K8s resources don’t have that, so the XR stays not ready.

**Fix:** Add **function-auto-ready** as a pipeline step. It runs resource-specific health checks (e.g. Deployment: `readyReplicas`, Service, ConfigMap) and marks the XR ready when all composed resources are ready.

This example already includes:
1. **function-auto-ready** in `functions.yaml` (install it).
2. A **mark-ready** step in `composition.yaml` that calls `function-auto-ready`.

**If you applied the example before this was added:**  
Apply the updated `functions.yaml` and `composition.yaml`, then wait for the new function to be healthy and for the pipeline to run again (or trigger a reconcile):

```bash
kubectl apply -f functions.yaml
kubectl get functions   # wait until function-auto-ready is INSTALLED and HEALTHY
kubectl apply -f composition.yaml
# Reconcile the XR (optional): kubectl annotate xapp my-app -n default crossplane.io/pause=false --overwrite
kubectl get xapp -n default -w
```

Once the Deployment has ready replicas (and the function has run), the XR should show **READY: True**.

## Cleanup

```bash
kubectl delete -f xr.yaml
kubectl delete -f composition.yaml
kubectl delete -f xrd.yaml
kubectl delete -f functions.yaml   # optional: remove the function
# optionally: kubectl delete -f rbac.yaml
```
