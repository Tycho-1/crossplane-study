# Example: Helm chart via Crossplane `provider-helm`

This folder holds **minimal samples** to complement **[guides/19_Configuration_From_Helm_Chart.md](../guides/19_Configuration_From_Helm_Chart.md)**.

It is **not** a full marketplace Configuration package — it shows the **managed resource** shape you typically **compose** from an XR.

## Prerequisites

1. **Crossplane** on a management cluster.  
2. **`provider-helm`** installed and HEALTHY.  
3. **`ProviderConfig`** for Helm with a kubeconfig `Secret` for the **target** cluster where the chart should run.  
4. (Optional) Target cluster already exists (e.g. from [configuration-aws-eks](../configuration-aws-eks/) EKS XR).

## Files

| File | Purpose |
|------|---------|
| **sample-release.yaml** | Example **`Release`** MR — adjust chart, repo, namespace, `providerConfigRef` to match your environment. |

## Apply order

```bash
# 1) Install provider-helm (see Marketplace / provider docs)
# 2) Create ProviderConfig + Secret (kubeconfig for target cluster)
kubectl apply -f sample-release.yaml
kubectl get release.helm.crossplane.io -A
```

## Next step: full Configuration

To expose a **custom API** (XR) instead of applying `Release` directly:

- Add an **XRD** + **Composition** that patches **`Release.spec`** from **`spec.parameters`** (see guide 19).

## References

- [provider-helm](https://github.com/crossplane-contrib/provider-helm)  
- [guide 19 – Configuration from Helm chart](../guides/19_Configuration_From_Helm_Chart.md)
