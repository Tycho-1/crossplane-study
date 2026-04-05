# Configuration from an existing Helm chart

This guide explains how to get **Crossplane** to drive **Helm chart–like** deployments while benefiting from **Crossplane reconciliation** (continuous desired state, GitOps-friendly APIs, composition). It also states what **cannot** be auto-magic.

**Related:** [Crossplane for Kubernetes applications](./09_Crossplane_For_Kubernetes_Applications.md) (composing in-cluster resources), [Configurations](./13_Configurations.md) (Cilium / `Release` mention).

---

## 1. What you are asking for (two different goals)

| Goal | What “Crossplane reconciliation” means | Typical approach |
|------|----------------------------------------|------------------|
| **A. Same chart, Crossplane-owned install** | Crossplane continuously ensures a **Helm release** matches `spec` (install/upgrade/delete, drift vs *release*). | **[provider-helm](https://github.com/crossplane-contrib/provider-helm)** — managed resource **`Release`** (`helm.crossplane.io/v1beta1`). |
| **B. Same resources as the chart, every object an MR** | Each Deployment/Service/… is its own managed resource with individual drift repair. | **Manual port** (or generators): expand chart → many **provider-kubernetes** `Object` MRs / raw YAML — **no** standard one-click converter from arbitrary charts. |

Most teams mean **A** when they say “Helm chart + Crossplane”: one **high-level API** (your XR) that composes a **`Release`** MR pointing at **chart + values**.

---

## 2. Recommended pattern: `provider-helm` `Release` inside a Composition

### Flow

```
Your XR (custom API)
    → Composition (function or P&T)
        → composed MR: helm.crossplane.io/Release
            → provider-helm reconciles Helm in the target cluster
                → same rendered resources Helm would create
```

- **Crossplane** reconciles the **`Release`** object (and your **XR**).
- **Helm** still renders the chart and applies objects into the workload cluster (same as `helm install`).
- **Drift:** provider-helm tracks **release** state; full per-object drift semantics are **Helm’s**, not Crossplane per-Pod, unless you use other tools.

### What you need

1. **`provider-helm`** installed (`Provider` package) and **HEALTHY**.
2. **`ProviderConfig`** for Helm pointing at the **target cluster** (kubeconfig `Secret` — often the same pattern as [configuration-aws-eks](../configuration-aws-eks/HOW-IT-WORKS.md) after EKS creates a kubeconfig `Secret`).
3. **XRD** — define your custom API (`spec.parameters`: chart name, repo, values, namespace, …).
4. **Composition** — output a **`Release`** whose `spec.forProvider` is filled from the XR (KCL, Go, Patch & Transform, CUE, etc.).
5. **Configuration package** (optional) — bundle XRD + Composition + `dependsOn` for `provider-helm`.

### Official references

- Provider: [crossplane-contrib/provider-helm](https://github.com/crossplane-contrib/provider-helm)  
- Marketplace: [provider-helm](https://marketplace.upbound.io/providers/crossplane-contrib/provider-helm)  
- MR type: **`Release`** — `helm.crossplane.io/v1beta1` (confirm with `kubectl explain release.helm.crossplane.io` on your cluster).

---

## 3. What you do **not** get automatically

| Expectation | Reality |
|-------------|---------|
| “Convert Chart.tgz → Crossplane Configuration with zero work” | **No** universal upstream tool; you **design** the XRD/Composition and map **values** to **`Release.spec`**. |
| “Every Pod is a separate Crossplane MR like AWS RDS” | **Not** with pure `Release`; Helm owns the rendered objects. Use **provider-kubernetes** `Object` or **Operators** if you need per-resource MRs. |
| “Helm hooks and ordering work identically” | **Mostly** yes **inside Helm**; edge cases exist — test your chart. |
| “Same as `helm install` without a cluster” | You still need a **target Kubernetes** for the release; Crossplane runs the **provider** from the **management** cluster. |

---

## 4. Building your own Configuration (outline)

1. **Fork or create** a repo with `upbound.yaml` / Crossplane project layout.  
2. Add **`dependsOn`**: `provider-helm` (and `provider-kubernetes` if you mix patterns).  
3. **Define XRD** — parameters you want platform users to set (chart URL, version, values YAML, target namespace).  
4. **Define Composition** — one composed resource: **`Release`**, with `spec.forProvider.chart`, `namespace`, `values`, `providerConfigRef.name` from XR.  
5. **`up project build`** / **`crossplane xpkg build`** → push OCI → install **Configuration** on the control plane.  
6. Users create **your XR**; Crossplane creates/updates the **`Release`**; Helm materializes the chart.

See the minimal samples in **[examples/configuration-from-helm-chart](../examples/configuration-from-helm-chart/)**. If you want the **same** chart materialized as **composed Kubernetes resources** (no Helm `Release` at runtime), see **[configuration-openforms](../configuration-openforms/)** — Helm is only used offline to render **reference** YAML for diffs and regeneration.

---

## 5. Alternative: `provider-kubernetes` `Object` (YAML blobs)

You can store **rendered** manifest YAML in **`Object`** MRs (or split per file). Crossplane reconciles **those** objects. Downsides: large specs, chart upgrades are harder, and you lose **Helm release** semantics unless you still run Helm elsewhere.

Use when you must avoid Helm at runtime or need very custom patching.

---

## 6. Relation to EKS + Helm `ProviderConfig`

[configuration-aws-eks](../configuration-aws-eks/) already composes a **Helm `ProviderConfig`** (and Kubernetes `ProviderConfig`) wired to the cluster kubeconfig `Secret`. That is the **credential hook** for a follow-up **`Release`** in the same Composition or a separate Configuration — see [CILIUM-TODO.md](../configuration-aws-eks/CILIUM-TODO.md) for the idea (add a Cilium **`Release`** after the cluster exists).

---

## 7. Summary

| Question | Answer |
|----------|--------|
| Can I wrap an existing Helm chart in a Crossplane Configuration? | **Yes** — compose a **`helm.crossplane.io/Release`** MR via **`provider-helm`**. |
| Do I get “all Crossplane reconciliation” for every Deployment inside the chart? | You get Crossplane reconciliation for the **Release** (and your **XR**). **In-cluster workloads** remain **Helm-managed** unless you model them separately. |
| Is it automatic from Chart.yaml alone? | **No** — you implement **XRD + Composition + values mapping**. |

For a **second topic** next to EKS, treat **Helm via `Release`** as the standard bridge; use **full MR expansion** only when you explicitly need per-resource control.

**Big picture:** Crossplane does **not** make Helm or Terraform unnecessary in general — they often **coexist** with different roles. See **[Helm, Terraform, and Crossplane: complement, not wholesale replacement](./20_Terraform_Modules_vs_Crossplane_Configurations.md#helm-terraform-and-crossplane-complement-not-wholesale-replacement)**.
