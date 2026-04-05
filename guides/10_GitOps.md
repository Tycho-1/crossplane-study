# GitOps with Crossplane

**GitOps** means using Git as the single source of truth for your cluster state. You store Crossplane manifests (XRDs, Compositions, XRs, MRs, ProviderConfigs) in a repo, and a GitOps controller (e.g. **Argo CD** or **Flux**) syncs that state into the cluster. Crossplane then reconciles XRs and MRs against the cloud or in-cluster resources.

This guide covers how to run Crossplane in a GitOps workflow and what to configure in Argo CD (and briefly Flux) so they work well together.

---

## Why GitOps with Crossplane?

- **Declarative infra and apps:** Everything (providers, XRDs, Compositions, claims) is in Git; changes go through PRs and history is auditable.
- **Same workflow everywhere:** Platform team owns XRDs/Compositions; dev teams (or other repos) own XR instances; both are synced by the same GitOps tool.
- **Drift detection:** The GitOps controller can show when cluster state has drifted from Git and optionally auto-sync.
- **Multi-cluster:** One or more clusters can sync from the same (or different) repos; Crossplane in each cluster reconciles MRs to the cloud.

---

## Argo CD with Crossplane

Argo CD syncs Kubernetes manifests from Git to the cluster. Crossplane creates **dynamic resources** (e.g. composed MRs, ProviderConfigUsage). For Argo CD to track and health-check these correctly, you need a few settings.

**Recommended:** Use **Argo CD 2.4.8 or later** with Crossplane. See [Configuring Crossplane with Argo CD](https://docs.crossplane.io/latest/guides/crossplane-with-argo-cd/).

### 1. Resource tracking: annotation-based

Crossplane creates child resources (composed resources, ProviderConfigUsage) that are not in Git. Argo CD must use **annotation-based** resource tracking so it can associate these with the right Application.

Edit the `argocd-cm` ConfigMap in the `argocd` namespace:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  application.resourceTrackingMethod: annotation
```

### 2. Health checks for Crossplane resources

Argo CD needs custom health logic for Crossplane CRs (Provider, Composition, XR, MR, ProviderConfig, etc.) so they show as Healthy/Progressing/Degraded instead of "Unknown".

Configure `resource.customizations` in `argocd-cm` with Lua health scripts for:

- `*.crossplane.io/*` – XRDs, Compositions, XRs, MRs, Provider, ProviderConfig, etc.
- `*.upbound.io/*` – same for Upbound/UXP.

The scripts typically check `status.conditions` (e.g. `Ready`, `Synced`) and treat resources with no status (e.g. ProviderConfig, ProviderConfigUsage) as Healthy when appropriate.

Full examples are in the [official guide](https://docs.crossplane.io/latest/guides/crossplane-with-argo-cd/#set-health-status). The built-in [Argo CD resource customizations](https://github.com/argoproj/argo-cd/tree/master/resource_customizations) already include `pkg.crossplane.io/Provider`; add the rest as in the Crossplane docs.

### 3. Hide ProviderConfigUsage from the UI

Providers create a **ProviderConfigUsage** for each MR. These are implementation details and can clutter the Argo CD UI. Exclude them:

```yaml
# In argocd-cm
data:
  resource.exclusions: |
    - apiGroups: ["*"]
      kinds: ["ProviderConfigUsage"]
```

### 4. Kubernetes client QPS (optional)

With many CRDs (e.g. many providers), the Argo CD application controller may hit API rate limits. Increase the client QPS, e.g.:

- Set env `ARGOCD_K8S_CLIENT_QPS=300` on the Argo CD application controller.

---

## Flux with Crossplane

**Flux** also uses Git as source of truth and reconciles Kubernetes resources. With Crossplane:

- Store XRDs, Compositions, Providers, ProviderConfigs, and XRs/MRs in a Git repo (or multiple repos).
- Use **Kustomization** or **HelmRelease** to apply them; respect **apply order** (e.g. XRD → Composition → XR).
- Flux does not create Crossplane’s dynamic children; it only applies what’s in Git. Crossplane then creates and manages composed resources and MRs.

Flux does not require annotation-based tracking like Argo CD; it tracks only the resources it applies. For health, you can use [custom health checks](https://fluxcd.io/flux/guides/helm-customization/#custom-health-checks) if needed.

Reference: [GitOps model for EKS with Crossplane and Flux](https://aws.amazon.com/blogs/containers/gitops-model-for-provisioning-and-bootstrapping-amazon-eks-clusters-using-crossplane-and-flux/).

---

## Directory structure (Git repo)

A common layout is to separate **platform** (Crossplane + APIs) from **workloads** (claims / XRs):

```
repo/
├── platform/                    # or crossplane/
│   ├── providers/               # Provider + ProviderConfig
│   │   ├── provider-aws.yaml
│   │   └── provider-config.yaml
│   ├── xrds/                    # CompositeResourceDefinitions
│   │   └── xbucket.yaml
│   └── compositions/
│       └── bucket-composition.yaml
├── workloads/                   # or environments/dev, prod
│   ├── dev/
│   │   └── buckets.yaml         # XBucket instances
│   └── prod/
│       └── buckets.yaml
```

- **Apply order:** Sync `platform/` first (providers, XRDs, Compositions), then `workloads/` (XRs). In Argo CD you can use multiple Applications with dependencies, or a single Application with a sync wave / ordering (e.g. Kustomize order, or Argo CD sync waves).
- **Kustomize:** Use a base + overlays or a single `kustomization.yaml` that lists resources in the right order.

---

## Apply order and sync

1. **Providers** and **ProviderConfig** (and Secrets for credentials).
2. **XRDs** (CompositeResourceDefinitions).
3. **Compositions** (reference the XRD).
4. **XRs / claims** (reference the Composition).

If the GitOps controller applies everything in one go, Crossplane may create an XR before the Composition exists; the XR will stay pending until the Composition is applied. Using **sync waves** (Argo CD) or **Kustomize order** (e.g. `config.kubernetes.io/depends-on`) helps avoid transient errors.

---

## Drift and reconciliation

- **Git → cluster:** The GitOps controller (Argo CD / Flux) reconciles Git state to the cluster. If someone changes a YAML in Git, the next sync updates the cluster.
- **Cluster → cloud (or in-cluster):** Crossplane reconciles XRs and MRs. If the cloud resource (or composed K8s resource) is changed outside Crossplane, the provider/controller will try to revert it to match the spec (provider-dependent behavior).

So you have two layers: GitOps keeps Crossplane resources in sync with Git; Crossplane keeps cloud and composed resources in sync with those Crossplane resources.

---

## References

- [Configuring Crossplane with Argo CD](https://docs.crossplane.io/latest/guides/crossplane-with-argo-cd/) – resource tracking, health, exclusions, QPS
- [Argo CD resource tracking](https://argo-cd.readthedocs.io/en/latest/user-guide/resource_tracking/)
- [GitOps for EKS with Crossplane and Argo CD](https://aws.amazon.com/blogs/containers/gitops-model-for-provisioning-and-bootstrapping-amazon-eks-clusters-using-crossplane-and-argo-cd/)
- [GitOps for EKS with Crossplane and Flux](https://aws.amazon.com/blogs/containers/gitops-model-for-provisioning-and-bootstrapping-amazon-eks-clusters-using-crossplane-and-flux/)
