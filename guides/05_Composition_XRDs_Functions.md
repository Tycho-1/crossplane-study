# Composition: XRDs, Compositions & Functions

Composition gives you **custom APIs** (Composite Resources, XRs) that create and manage **multiple resources** (MRs, Deployments, Services, etc.) through a single object. In v2, composed resources can be **any** Kubernetes resource, not only Crossplane MRs.

**Using Composition for Kubernetes applications (v2):** You can use Crossplane purely for **normal K8s applications** (Deployments, Services, ConfigMaps)—no cloud MRs. Define an XRD (e.g. `XApp`) and a Composition whose function pipeline outputs Deployment + Service; users create one XR and get the app. See [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md) and [examples/05-app-with-deployment](../examples/05-app-with-deployment/).

---

## Building blocks

| Resource | Role |
|----------|------|
| **XRD** (Composite Resource Definition) | Defines the **API**: group, kind, schema (openAPI or CEL). |
| **Composition** | Defines **how** that API is implemented: pipeline of **functions** that produce composed resources. |
| **XR** (Composite Resource) | An **instance** of the API; creating one triggers the Composition and creates the composed resources. |
| **Composition Function** | A step in the pipeline that takes observed/desired state + input and returns desired composed resources (and optionally readiness). |

Flow: **User creates XR** → Crossplane selects a **Composition** for that XR kind → runs the **function pipeline** → **applies** the composed resources.

---

## Composite Resource Definition (XRD)

XRD defines the **custom API** (group, kind, versions, schema). In v2 you can set **scope** to `Namespaced` (default) or `Cluster`.

**Minimal example (v2 API):**

```yaml
apiVersion: apiextensions.crossplane.io/v2
kind: CompositeResourceDefinition
metadata:
  name: xbuckets.example.org
spec:
  scope: Namespaced
  group: example.org
  names:
    kind: XBucket
    plural: xbuckets
  versions:
    - name: v1alpha1
      served: true
      referenceable: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                region:
                  type: string
                  default: us-east-2
              required: [region]
```

Legacy v1 XRD API is still supported; it defaults to cluster-scoped and Claims.

---

## Compositions and the function pipeline

A **Composition** ties an XR **kind** to a **pipeline** of composition functions. Each step calls one **Function** (e.g. Patch & Transform, CUE, Python) and can receive **input**. The pipeline produces the **desired** set of composed resources.

### Mode: Pipeline

Always use `mode: Pipeline` and a `pipeline` list of steps:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: bucket-pt
spec:
  compositeTypeRef:
    apiVersion: example.org/v1alpha1
    kind: XBucket
  mode: Pipeline
  pipeline:
    - step: create-bucket
      functionRef:
        name: function-patch-and-transform
      input:
        apiVersion: pt.fn.crossplane.io/v1beta1
        kind: Resources
        resources:
          - name: bucket
            base:
              apiVersion: s3.aws.m.upbound.io/v1beta1
              kind: Bucket
              spec:
                forProvider:
                  region: us-east-2
            patches:
              - type: FromCompositeFieldPath
                fromFieldPath: spec.region
                toFieldPath: spec.forProvider.region
```

- **compositeTypeRef** must match the XRD’s group/kind (and version).
- **functionRef.name** must match an installed **Function** (e.g. `function-patch-and-transform`).
- **input** is function-specific (e.g. Patch & Transform uses `Resources` with `base` + `patches`).

### Matching Compositions to XRs

- **Default:** Crossplane picks a Composition whose `compositeTypeRef` matches the XR.
- **Explicit:** On the XR, set `spec.crossplane.compositionRef.name` (and optionally `compositionRevisionRef`).
- **Selector:** Use `spec.crossplane.compositionSelector.matchLabels` to choose by label.

### Composition revisions

Each change to a Composition creates a **CompositionRevision**. XRs can:

- Use the latest revision by default (`compositionUpdatePolicy: Automatic`).
- Pin to a revision with `compositionRevisionRef.name` and `compositionUpdatePolicy: Manual`.

---

## Function: Patch & Transform

**function-patch-and-transform** is the standard way to define composed resources with YAML and patches (similar to the old embedded P&T).

Install:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.10.1
```

Input kind: `pt.fn.crossplane.io/v1beta1`, `Resources`. Each entry has:

- **name:** logical name of the composed resource.
- **base:** base resource spec (e.g. Bucket).
- **patches:** list of patches (FromCompositeFieldPath, ToCompositeFieldPath, Combine, etc.).

**Example patch – pass XR spec to MR:**

```yaml
patches:
  - type: FromCompositeFieldPath
    fromFieldPath: spec.region
    toFieldPath: spec.forProvider.region
```

See [Function Patch and Transform](https://docs.crossplane.io/latest/guides/function-patch-and-transform/) for full patch types.

---

## Other functions (CUE, KCL, Python, Go template)

- **CUE:** [function-cue](https://github.com/crossplane-contrib/function-cue) – express templates in CUE.
- **KCL:** [function-kcl](https://github.com/crossplane-contrib/function-kcl) – use KCL for logic.
- **Python:** [function-python](https://github.com/crossplane-contrib/function-python) – script in Python.
- **Go templating:** [function-go-templating](https://github.com/crossplane-contrib/function-go-templating) – Helm-like YAML templates.

You can chain steps: e.g. first step creates resources (CUE/P&T), second step is **function-auto-ready** to derive readiness from composed resources.

---

## Required resources (bootstrap and dynamic)

Functions can **read** existing cluster resources (ConfigMaps, other CRs) to build desired state.

- **Bootstrap:** declare `requirements.requiredResources` in the Composition step; Crossplane fetches them before calling the function.
- **Dynamic:** function returns a request for required resources in its response; Crossplane fetches and calls the function again (up to a limit). Use when the resource name/namespace depends on XR state.

---

## Granting Crossplane access to composed resources

- Crossplane already has access to MRs and XRs from providers and XRDs.
- For **any other** API (e.g. `postgresql.cnpg.io`, core `Deployment`/`Service`), add a **ClusterRole** that **aggregates** to Crossplane:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-api:aggregate-to-crossplane
  labels:
    rbac.crossplane.io/aggregate-to-crossplane: "true"
rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "services"]
    verbs: ["*"]
```

---

## Testing without a cluster: crossplane render

You can **render** the result of a Composition locally with the **Crossplane CLI** (needs Docker for running functions):

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

Output: the XR plus all composed resources as YAML. Use this to iterate on Compositions and functions without applying.

---

## v2 specifics

- **spec.crossplane:** All composition machinery (compositionRef, compositionRevisionRef, resourceRefs, compositionUpdatePolicy) lives under **spec.crossplane** on the XR.
- **Namespaced XRs:** Composed resources are created in the XR’s namespace when they are namespaced.
- **Compose any resource:** Include Deployments, Services, or third-party CRs in your function output; ensure RBAC is in place.

---

## Next

- [Operations](./06_Operations.md) – CronOperation, WatchOperation.
- [Packages & CLI](./07_Packages_CLI.md) – Installing functions, `crossplane render`.
- Practice: [examples/03-xrd-composition-bucket](../examples/03-xrd-composition-bucket/), [examples/04-patch-and-transform](../examples/04-patch-and-transform/), [examples/05-app-with-deployment](../examples/05-app-with-deployment/).
