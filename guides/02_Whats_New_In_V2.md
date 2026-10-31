# What's New in Crossplane v2

Crossplane v2 makes the platform **more intuitive** and **less opinionated**, and better suited for **applications** as well as infrastructure. **Applications are now first-class:** you can use Crossplane not only for cloud resources (S3, RDS, EKS) but also for **normal Kubernetes applications** (Deployments, Services, ConfigMaps) in the same cluster. See [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md).

**Four major changes:**

1. **Composite resources (XRs) are namespaced by default.**
2. **Managed resources (MRs) are namespaced.**
3. **Compositions can compose any Kubernetes resource** (not only Crossplane MRs).
4. **Operations** for declarative day-two workflows (run-once, cron, watch).

---

## 1. Namespaced Composite Resources (XRs)

- In v2, XRs are **namespaced** by default.
- A namespaced XR composes resources **in its own namespace**.
- Crossplane “machinery” lives under **`spec.crossplane`**, so user-facing fields are clearly separated.

**Example (v2 namespaced XR):**

```yaml
apiVersion: example.crossplane.io/v1
kind: App
metadata:
  namespace: default
  name: my-app
spec:
  image: nginx
  crossplane:
    compositionRef:
      name: app-kcl
    compositionRevisionRef:
      name: app-kcl-41b6efe
    resourceRefs:
      - apiVersion: apps/v1
        kind: Deployment
        name: my-app-9bj8j
      - apiVersion: v1
        kind: Service
        name: my-app-bflc4
```

**XRD scope:** In the v2 XRD API (`apiextensions.crossplane.io/v2`), use `scope: Namespaced` (default) or `scope: Cluster`. With v1 XRD API, you get legacy cluster-scoped XRs and Claims.

- **No more Claims** for the new model: namespaced XRs are the primary abstraction; cluster-scoped XRs are optional.

---

## 2. Namespaced Managed Resources (MRs)

- In v2, **managed resources are namespaced** (e.g. S3 Bucket, RDS instance).
- Enables **fully namespaced** compositions: one namespace can contain an XR and all its composed resources (MRs, Deployments, Services, etc.).
- **RBAC:** You can scope who can create which MRs by namespace and role.

**Example (v2 namespaced Bucket):**

```yaml
apiVersion: s3.aws.m.upbound.io/v1beta1
kind: Bucket
metadata:
  namespace: default
  generateName: my-bucket-
spec:
  forProvider:
    region: us-east-2
```

**Compatibility:** v2 still supports **cluster-scoped MRs** (legacy). New provider releases support both; cluster-scoped MRs will be deprecated later.

**Managed Resource Definitions (MRDs):** v2 adds **selective activation** of provider MRs so you only install the CRDs you need.

---

## 3. Compose Any Resource

- Compositions are **no longer limited** to Crossplane MRs.
- An XR can compose:
  - Crossplane **MRs** (e.g. RDSInstance, Bucket),
  - **Kubernetes built-in** resources (Deployment, Service, ConfigMap),
  - **Third-party CRs** (e.g. CloudNativePG Cluster, Cluster API Cluster).

So you can define an **“App”** XR that creates a Deployment, Service, database MR, and monitoring in one go.

**Access:** Crossplane can already manage resources from providers and XRDs. For **other** resource types (e.g. `postgresql.cnpg.io/Cluster`), you must grant Crossplane RBAC (e.g. ClusterRole that aggregates to Crossplane’s ClusterRole).

**Example – ClusterRole for CloudNativePG:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cnpg:aggregate-to-crossplane
  labels:
    rbac.crossplane.io/aggregate-to-crossplane: "true"
rules:
  - apiGroups: ["postgresql.cnpg.io"]
    resources: ["clusters"]
    verbs: ["*"]
```

---

## 4. Operations (Day-Two Workflows)

**Operations** run **function pipelines to completion** (like a Job), instead of continuously reconciling resources. Use cases: cert checks, rolling upgrades, scheduled maintenance, reacting to resource changes.

**Modes:**

- **Operation** – run once to completion.
- **CronOperation** – run on a schedule (e.g. daily cert check).
- **WatchOperation** – run when selected resources change.

**Example – CronOperation (alpha):**

```yaml
apiVersion: ops.crossplane.io/v1alpha1
kind: CronOperation
metadata:
  name: cert-monitor
spec:
  schedule: "0 6 * * *"   # Daily at 6 AM
  operationTemplate:
    spec:
      mode: Pipeline
      pipeline:
        - step: check-certificates
          functionRef:
            name: crossplane-contrib-function-python
      # function checks SSL certs and reports status
```

**Status:** Operations are **alpha** in v2.

---

## Backward Compatibility and Breaking Changes

**Most v1 users can upgrade without breaking changes** if they avoid deprecated/alpha features.

**Removed in v2:**

- **Native patch and transform** in Compositions → use **composition functions** (e.g. function-patch-and-transform).
- **ControllerConfig** → use **DeploymentRuntimeConfig**.
- **External secret stores** (alpha, unmaintained).
- **Composite resource connection details** → compose your own `Secret` (see [connection details composition](https://docs.crossplane.io/latest/guides/connection-details-composition/)).
- **Default package registry** → you must use **fully qualified package URLs** (e.g. `xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0`).

**Before upgrading:**

- Ensure all packages use **fully qualified image URLs**.
- Run `kubectl get pkg` and fix any that don’t.
- Follow the [Upgrade to Crossplane v2](https://docs.crossplane.io/latest/guides/upgrade-to-crossplane-v2/) guide.

**Legacy v1 XRs/MRs:** v2 still supports them; existing Compositions need minor updates for v2-style XRs/MRs (see upgrade guide).

---

## Summary Table

| Area | v1 | v2 |
|------|----|----|
| XR scope | Cluster (Claims in namespaces) | Namespaced (default), optional Cluster |
| MR scope | Cluster | Namespaced |
| Composed resources | Crossplane MRs (and XRs) | Any K8s resource (MR, Deployment, CRs) |
| Claims | Yes | No (use namespaced XRs) |
| XR “machinery” | Mixed in spec | Under `spec.crossplane` |
| Day-two workflows | Ad hoc | Operations (CronOperation, etc.) |
| Composition | P&T or functions | Functions (e.g. P&T, CUE, KCL, Python) |

---

## Next

- [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md) – Using Crossplane for normal K8s apps (Deployments, Services, etc.).
- [Installation & Setup](./03_Installation_Setup.md) – Install Crossplane and set feature flags.
- [Composition (XRDs, Compositions, Functions)](./05_Composition_XRDs_Functions.md) – Define APIs and implement them with functions.
