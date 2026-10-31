# Configurations

A **Configuration** is a Crossplane package that bundles your platform's building blocks — XRDs, Compositions, and Function dependencies — into a single versioned, installable OCI image. It is the standard way to ship and distribute a platform API.

---

## What a Configuration bundles

```
my-configuration (OCI image)
├── crossplane.yaml          ← package metadata + dependencies
├── xrd.yaml                 ← one or more XRDs  (your custom API)
├── composition.yaml         ← one or more Compositions (how to fulfil the API)
└── (no XRs, no Providers)   ← XRs are created by users; Providers are separate packages
```

| Included | Not included |
|---|---|
| XRDs | XR instances (those are created by the user) |
| Compositions | Providers (declared as dependencies, not bundled) |
| Function dependencies (declared) | ProviderConfigs (cluster-specific credentials) |
| | MRDs (Managed Resource Definitions) — come from Providers, not from the Configuration |

**Note:** Composed resources are often **MRs** (Managed Resources) like `Bucket`, `Table`, `Cluster` — defined by **MRDs** (Managed Resource Definitions) that Providers install. See [Managed Resources & Providers](./04_Managed_Resources_Providers.md) and [MRD in Packages](./14_Packages_And_xpkg_CLI.md#mrd-managed-resource-definition).

---

## How it relates to the other components

```
┌─────────────────────────────────────────────────────────┐
│  Configuration package (installed once by platform team) │
│                                                          │
│   XRD  ──defines──▶  XPlatform (custom API kind)        │
│   Composition  ────▶  pipeline: function-p&t             │
│                              │                           │
└──────────────────────────────┼───────────────────────────┘
                               │ declares dependency on
                               ▼
                    Function: function-patch-and-transform
                    Provider: provider-aws-s3  (auto-installed)

                    User creates:
                    XR (XPlatform instance)
                         │
                         ▼
                    Crossplane runs the Composition pipeline
                         │
                         ▼
                    Composed resources: Bucket, DynamoDB, IAM Policy
```

In short:
- The **Configuration** ships the API definition and the implementation.
- The **user** only ever creates the XR — they never touch the XRD or Composition directly.
- **Providers** and **Functions** are declared as dependencies inside the Configuration; Crossplane installs them automatically when the Configuration is installed.

---

## The crossplane.yaml file

Every Configuration package must have a `crossplane.yaml` at the root. It declares the package type, metadata, and dependencies.

```yaml
apiVersion: meta.pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
  annotations:
    meta.crossplane.io/maintainer: "your-team@example.org"
    meta.crossplane.io/description: "Platform API: XPlatform (Bucket + DynamoDB + IAM)"
spec:
  crossplane:
    version: ">=2.0.0"           # minimum Crossplane version required
  dependsOn:
    - function: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform
      version: ">=v0.10.0"
    - provider: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3
      version: ">=v2.0.0"
    - provider: xpkg.upbound.io/upbound/provider-aws-dynamodb
      version: ">=v2.0.0"
    - provider: xpkg.upbound.io/upbound/provider-aws-iam
      version: ">=v2.0.0"
```

When you install this Configuration, Crossplane automatically installs all `dependsOn` Providers and Functions.

---

## Provider packages and Kubernetes RBAC

**Who creates `ClusterRole`, `ClusterRoleBinding`, and the provider `ServiceAccount`?**

- **Not** your Configuration package. Configurations bundle XRDs, Compositions, and metadata — they do **not** contain the RBAC for AWS/GCP/Azure **provider controllers**.
- **Provider** packages do. Each **Provider** OCI image embeds Kubernetes RBAC (ClusterRoles, bindings, ServiceAccounts). When you `kubectl apply` a `Provider` CR (or Crossplane installs one via a Configuration’s `dependsOn`), **Crossplane’s package controller** unpacks the image and **creates** those objects in the cluster.
- **Crossplane core** (Helm install) has its own RBAC for the Crossplane and RBAC manager — separate from per-provider rules.

If a provider pod logs `forbidden: cannot list ... providerconfigs.aws.m.upbound.io`, the provider’s ClusterRole is missing or stale; **reinstall or upgrade the Provider** so RBAC is reapplied — you typically do not patch RBAC by hand for official packages.

**See also:** [configuration-aws-eks README — Who creates ClusterRole…](../configuration-aws-eks/README.md#who-creates-clusterrole-rolebinding-and-serviceaccount) (concrete example + debug commands), [Packages & xpkg CLI](./14_Packages_And_xpkg_CLI.md), [Managed Resources & Providers](./04_Managed_Resources_Providers.md).

---

## Full example: building your own Configuration

The following recreates example 07 (Bucket + DynamoDB + IAM Policy) as a proper Configuration package.

### Directory layout

```
my-platform/
├── crossplane.yaml      ← package metadata + dependencies
├── xrd.yaml             ← XPlatform API definition
└── composition.yaml     ← how XPlatform is implemented
```

### crossplane.yaml

```yaml
apiVersion: meta.pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
  annotations:
    meta.crossplane.io/description: "XPlatform: S3 Bucket + DynamoDB Table + IAM Policy"
spec:
  crossplane:
    version: ">=2.0.0"
  dependsOn:
    - function: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform
      version: ">=v0.10.0"
    - provider: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3
      version: ">=v2.0.0"
    - provider: xpkg.upbound.io/upbound/provider-aws-dynamodb
      version: ">=v2.0.0"
    - provider: xpkg.upbound.io/upbound/provider-aws-iam
      version: ">=v2.0.0"
```

### xrd.yaml

```yaml
apiVersion: apiextensions.crossplane.io/v2
kind: CompositeResourceDefinition
metadata:
  name: xplatforms.example.org
spec:
  scope: Namespaced
  group: example.org
  names:
    kind: XPlatform
    plural: xplatforms
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
                storageName:
                  type: string
                tableName:
                  type: string
                iamPolicyName:
                  type: string
              required: [region, storageName, tableName, iamPolicyName]
```

### composition.yaml

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: xplatform
spec:
  compositeTypeRef:
    apiVersion: example.org/v1alpha1
    kind: XPlatform
  mode: Pipeline
  pipeline:
    - step: create-resources
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
              - type: FromCompositeFieldPath
                fromFieldPath: spec.storageName
                toFieldPath: metadata.name
                policy:
                  fromFieldPath: Required

          - name: dynamodb-table
            base:
              apiVersion: dynamodb.aws.m.upbound.io/v1beta1
              kind: Table
              spec:
                forProvider:
                  region: us-east-2
                  billingMode: PAY_PER_REQUEST
                  hashKey: LockID
                  attribute:
                    - name: LockID
                      type: S
            patches:
              - type: FromCompositeFieldPath
                fromFieldPath: spec.region
                toFieldPath: spec.forProvider.region
              - type: FromCompositeFieldPath
                fromFieldPath: spec.tableName
                toFieldPath: metadata.name
                policy:
                  fromFieldPath: Required

          - name: iam-policy
            base:
              apiVersion: iam.aws.m.upbound.io/v1beta1
              kind: Policy
              spec:
                forProvider:
                  description: "S3 read/write access"
                  policy: |
                    {
                      "Version": "2012-10-17",
                      "Statement": [{
                        "Effect": "Allow",
                        "Action": ["s3:PutObject","s3:GetObject","s3:DeleteObject"],
                        "Resource": "arn:aws:s3:::*"
                      }]
                    }
            patches:
              - type: FromCompositeFieldPath
                fromFieldPath: spec.iamPolicyName
                toFieldPath: metadata.name
                policy:
                  fromFieldPath: Required
```

---

## Build and push the Configuration image

Use the Crossplane CLI to build the package into an OCI image and push it to a registry.

```bash
# Build the OCI image from the current directory
crossplane xpkg build \
  --package-root=my-platform/ \
  --output=my-platform.xpkg

# Push to a registry (e.g. Docker Hub, GitHub Container Registry, Upbound)
crossplane xpkg push \
  --package=my-platform.xpkg \
  docker.io/yourorg/my-platform:v0.1.0
```

The result is a standard OCI image — any OCI-compatible registry works (Docker Hub, GHCR, ECR, Upbound Marketplace).

---

## Install the Configuration in a cluster

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
spec:
  package: docker.io/yourorg/my-platform:v0.1.0
```

```bash
kubectl apply -f configuration-install.yaml

# Watch installation progress (Crossplane auto-installs all dependsOn)
kubectl get configurations.pkg.crossplane.io
kubectl get providers.pkg.crossplane.io
kubectl get functions.pkg.crossplane.io
```

Once `INSTALLED=True` and `HEALTHY=True`, the `XPlatform` kind is available in the cluster.

---

## User workflow after installation

The user never touches the XRD or Composition. They just create an XR:

```yaml
apiVersion: example.org/v1alpha1
kind: XPlatform
metadata:
  namespace: default
  name: my-platform
spec:
  region: eu-central-1
  storageName: my-platform-storage
  tableName: my-platform-lock-table
  iamPolicyName: my-platform-s3-policy
```

```bash
kubectl apply -f xr.yaml
kubectl get xplatforms -n default
kubectl get bucket
kubectl get table.dynamodb
```

---

## Versioning and upgrades

Configurations follow standard OCI image versioning. To upgrade:

```bash
# Update the package field to the new version and re-apply
kubectl patch configuration my-platform \
  --type merge \
  -p '{"spec":{"package":"docker.io/yourorg/my-platform:v0.2.0"}}'

# Crossplane creates a new ConfigurationRevision and activates it
kubectl get configurationrevisions.pkg.crossplane.io
```

Existing XRs keep working during the upgrade. Crossplane reconciles them against the new Composition once the new revision is active.

---

## Summary: who owns what

| Who | What they manage |
|---|---|
| **Platform team** | Builds and publishes the Configuration (XRD + Composition + crossplane.yaml) |
| **Cluster admin** | Installs the Configuration + ProviderConfig (credentials) once |
| **Developer / user** | Creates XR instances — never sees XRDs or Compositions |

This separation is the core value of Configurations: the platform team ships an opinionated, versioned API; users consume it without needing to know how it works internally.

---

## Comparison with Terraform modules

The analogy between Crossplane Configurations and Terraform modules is very close.

| | Terraform Module | Crossplane Configuration |
|---|---|---|
| **Purpose** | Bundle reusable infrastructure logic | Bundle reusable platform API (XRDs + Compositions) |
| **Input** | `variables.tf` | XRD schema (`spec.properties`) |
| **Implementation** | `main.tf` resources | Composition pipeline (managed resources) |
| **Versioning** | Git tag / registry version | OCI image tag (`v0.1.0`) |
| **Distribution** | Terraform Registry | OCI registry / Upbound Marketplace |
| **User calls it with** | `module "x" { source = "..." }` | `kubectl apply -f xr.yaml` |
| **Dependencies** | `required_providers` | `crossplane.yaml` `dependsOn` |

The mental model maps almost 1:1:
- `variables.tf` = XRD schema
- `main.tf` = Composition
- `module source + version` = Configuration OCI image + tag
- `terraform apply` = `kubectl apply -f xr.yaml`

**Discovering XR parameters (Terraform variables equivalent):** Terraform modules expose `variable` blocks; Crossplane Configurations expose parameters via the **XRD schema**. After installing a Configuration, run `kubectl get xrd <name> -o yaml` and inspect `spec.versions[].schema.openAPIV3Schema.properties.spec.properties.parameters`. Good packages also document parameters in their README or Marketplace page. See [configuration-aws-eks/XR-PARAMETERS.md](../configuration-aws-eks/XR-PARAMETERS.md) for a full example.

### Key differences

**1. Reconciliation loop vs one-shot**
Terraform runs once and is done (unless you re-run it). Crossplane continuously reconciles — if someone deletes the S3 bucket manually, Crossplane recreates it automatically. The desired state is always enforced.

**2. State storage**
Terraform stores state in a file/backend and has the drift detection problem. Crossplane has no state file — the XR object in Kubernetes *is* the state, and the real-world cloud resource is always reconciled against it.

**3. Kubernetes-native**
Crossplane lives entirely in Kubernetes. You use `kubectl`, RBAC, and GitOps (Flux/ArgoCD) natively. Terraform needs its own access management, pipelines, and state backend.

**4. Who triggers it**
In Terraform a human or CI pipeline runs `terraform apply`. In Crossplane a developer just creates an XR object — the platform reconciles it automatically and continuously.

---

## Best practices when working with Configurations

**Treat `my-platform/` like a Terraform module repo:**
- One git repo per Configuration
- Semantic versioning on tags (`v0.1.0`, `v1.0.0`)
- CI builds and pushes the OCI image on every tag
- Users pin to a version in their Configuration install manifest

**Separate concerns:**
```
platform-team/my-platform/    ← Configuration repo (XRD + Composition) — owned by platform team
app-team/my-app-infra/        ← XR instances only — owned by app teams (like a Terraform root module)
```

**Use GitOps for XR instances:**
- Store XR YAML in git
- Flux/ArgoCD applies them — same pattern as Terraform workspaces managed via GitOps

**Never put credentials in the Configuration** — `ProviderConfig` (credentials) always lives separately in the cluster, never inside the package.

---

## Real-world example: dot-kubernetes Configuration

[devops-toolkit/dot-kubernetes](https://marketplace.upbound.io/configurations/devops-toolkit/dot-kubernetes/v2.0.21) is a good public reference for how a real Configuration repo is structured. It provides fully operational Kubernetes clusters on EKS, GKE, and AKS from a single XR.

Install manifest (from the [Upbound Marketplace](https://marketplace.upbound.io/configurations/devops-toolkit/dot-kubernetes/v2.0.21)):

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: devops-toolkit-dot-kubernetes
spec:
  package: xpkg.upbound.io/devops-toolkit/dot-kubernetes:v2.0.21
```

### How it works: one XRD, multiple Compositions, label-based selection

This is the most important pattern to understand. There is **one `Cluster` XRD** but **three Compositions** — one per cloud provider:

| Composition name | Label | Target |
|---|---|---|
| `cluster-aws` | `provider: aws` | EKS (default) |
| `cluster-google` | `provider: google` | GKE |
| `cluster-azure` | `provider: azure` | AKS |

All three Compositions point to the same XRD kind (`devopstoolkit.ai/v2 Cluster`). The XRD declares a **default**:

```yaml
spec:
  defaultCompositionRef:
    name: cluster-aws   # used if the user doesn't specify a provider
```

The user chooses the cloud by setting a label selector on the XR:

```yaml
apiVersion: devopstoolkit.ai/v2
kind: Cluster
metadata:
  name: my-cluster
  namespace: default
spec:
  crossplane:
    compositionSelector:
      matchLabels:
        provider: google   # picks cluster-google Composition
  parameters:
    nodeSize: medium
    minNodeCount: 2
```

Crossplane matches the label against the Compositions' labels and runs the correct one. If you omit the selector entirely, it falls back to `cluster-aws`.

### Two pipeline steps per Composition

Each Composition has exactly **two steps** in its pipeline:

```
Step 1: "google" / "aws" / "azure"
    → Creates the cloud-specific infrastructure:
      EKS Cluster + NodeGroups + IAM roles  (aws)
      GKE Cluster + NodePool               (google)
      AKS Cluster + NodePool               (azure)

Step 2: "apps"  (identical across all three Compositions)
    → Installs apps into the new cluster via Helm/Kubernetes provider:
      Cilium, ArgoCD, Crossplane, Traefik, KEDA, Prometheus, ...
      (only the ones the user enabled via spec.parameters.apps.*)
```

Step 2 is completely **provider-agnostic** — it uses `provider-helm` and `provider-kubernetes` to install Helm charts into the freshly created cluster, regardless of whether it's EKS, GKE, or AKS. The only provider-specific bits in step 2 are small conditional blocks (e.g. Cilium needs different config flags per cloud).

### KCL for conditional logic

Instead of `function-patch-and-transform`, they use `function-kcl` — a real programming language embedded in the Composition. This allows `if/else`, loops, and computed values that are impossible in pure YAML patches.

Key patterns in the KCL code:

```python
# Read the XR fields
oxr = option("params").oxr

# Conditional resource creation — only create if user enabled it
if oxr.spec.parameters?.apps?.argocd?.enabled:
    chart { _name = "argo-cd", ... }

# Map a human-readable size to an actual machine type
if oxr.spec.parameters.nodeSize == "small":
    machineType = "e2-standard-2"
elif oxr.spec.parameters.nodeSize == "medium":
    machineType = "e2-standard-4"
else:
    machineType = "e2-standard-16"
```

This is why the compositions.yaml is 213KB — the KCL source for all three clouds + all optional apps is very large, but it would be even more verbose (and unmaintainable) in plain YAML patches.

### Status writeback

The KCL also writes back into the XR's `status` fields so the user can see cluster health directly from the XR object:

```python
status.clusterName = oxr.metadata.name
status.controlPlaneStatus = ocds[oxr.metadata.name + "-cluster"].Resource.status.conditions[0].reason
status.nodePoolStatus = ocds[oxr.metadata.name + "-nodepool"].Resource.status.conditions[0].reason
```

Which is why `kubectl get clusters` shows these columns:

```
NAME         CLUSTERNAME   CONTROLPLANE   NODEPOOL
my-cluster   my-cluster    Available      Available
```

### `Usage` resources: ordering dependencies

Inside the KCL you see many `Usage` resources (`protection.crossplane.io/v1beta1`). These declare ordering — "don't create resource B until resource A is healthy":

```python
# Don't install the Helm ProviderConfig until the GKE Cluster exists
usage { _nameSuffix = "cilium", _kind = "Release" }
```

Without `Usage`, Crossplane would try to create all composed resources in parallel. The cluster might not exist yet when Crossplane tries to install Helm charts into it, causing failures. `Usage` solves this.

### How its repo maps to your my-platform/

The source is at [github.com/vfarcic/crossplane-kubernetes](https://github.com/vfarcic/crossplane-kubernetes). The files that end up in the OCI image all live in the `package/` subfolder — that is the folder passed to `crossplane xpkg build`:

| Your `my-platform/` | Their `package/` | Purpose |
|---|---|---|
| `crossplane.yaml` | `package/crossplane.yaml` | Package metadata + `dependsOn` |
| `xrd.yaml` | `package/definition.yaml` | XRD (different filename, same thing) |
| `composition.yaml` | `package/compositions.yaml` | Compositions (all in one large file) |

### What lives outside the package folder

The rest of the repo is tooling and is **not** bundled into the OCI image:

```
crossplane-kubernetes/
├── package/               ← built into the OCI image
│   ├── crossplane.yaml
│   ├── definition.yaml    ← XRD
│   └── compositions.yaml  ← 213KB — generated by KCL, not written by hand
│
├── kcl/                   ← KCL source that generates compositions.yaml
├── examples/              ← example XR instances (equivalent to your xr.yaml)
├── providers/             ← provider install manifests
├── tests/                 ← automated tests (chainsaw)
├── .github/workflows/     ← CI: runs tests, builds & pushes OCI image on git tag
└── Justfile               ← task runner (build, test, publish commands)
```

The key difference from `my-platform/` is that their `compositions.yaml` is generated from **KCL source** (`kcl/` folder) rather than written as raw YAML. This makes sense at scale — 213KB of YAML is unmanageable by hand. For simpler platforms, writing YAML directly (as in `my-platform/`) is perfectly valid.

---

## Package vs Module — clarifying the terminology

In Crossplane, **"package"** is a broader term than **"module"** in Terraform. It refers to any versioned OCI image that Crossplane can install. There are three package types:

| Crossplane package type | Terraform equivalent | Purpose |
|---|---|---|
| **Configuration** | **Module** | Bundles reusable platform logic: XRDs + Compositions |
| **Provider** | `required_providers` block | Installs the controller + CRDs for an external system (AWS, GCP, ...) |
| **Function** | No equivalent | A composition pipeline step (patch-and-transform, KCL, go-templating, ...) |

So:
- **Configuration = Terraform module** ✓
- **Package ≠ Terraform module** — "package" is the generic term for all three types above

All three share the same mechanics: versioned OCI image, installed via a CR (`Provider`, `Function`, or `Configuration`), managed by revisions.

---

## Adoption patterns in companies

Companies using Crossplane in production fall into three camps. The pattern mirrors the evolution of Terraform adoption.

### Camp 1 — Raw XRDs/Compositions (most common today)

The platform team writes XRDs and Compositions directly in a git repo and applies them via GitOps (Flux/ArgoCD). No Configuration package is built or published. This is the equivalent of copy-pasting Terraform code instead of extracting modules — it works for small teams but doesn't scale.

```
platform-repo/
├── xrds/
│   └── xplatform.yaml
├── compositions/
│   └── xplatform-composition.yaml
└── providers/
    └── provider-aws-s3.yaml
```

### Camp 2 — Configurations as internal packages (recommended)

The platform team builds a Configuration OCI image and publishes it to an internal registry (ECR, GHCR, Artifactory). App teams install it and only ever create XR instances — they never touch XRDs or Compositions. This is the direct equivalent of the Terraform module pattern.

```
platform-team/my-platform/   →  builds  →  docker.io/mycompany/my-platform:v1.2.0
                                               ↓ installed in cluster
app-team-a/infra/            →  only XR YAML files
app-team-b/infra/            →  only XR YAML files
```

### Camp 3 — Upbound Spaces / managed control planes (enterprise scale)

Large enterprises use [Upbound Spaces](https://docs.upbound.io/spaces/) — each team gets its own managed Crossplane control plane. The platform team publishes Configurations centrally; teams consume them from their own isolated control plane. Most scalable model, requires Upbound's commercial offering.

---

## Is using Configurations best practice?

**Yes — Configurations are the official Crossplane best practice**, for the same reasons Terraform modules are best practice:

- **Versioning** — teams pin to `v1.2.0`; the platform team releases `v1.3.0` without breaking anyone
- **Abstraction** — app teams never touch XRDs or Compositions
- **Reuse** — one Configuration, many teams and clusters
- **Dependency management** — `dependsOn` auto-installs the correct Providers and Functions

### Why many companies are still in Camp 1

- Crossplane v2 is relatively new — teams are still learning the ecosystem
- Building and publishing OCI images adds CI/CD complexity that teams avoid early on
- Documentation tooling is less mature than Terraform's (no auto-generated docs from the XRD schema)
- Small teams with a single control plane don't feel the pain of Camp 1 yet

The pattern that consistently emerges is: start with raw Compositions, feel the pain of copy-paste and missing versioning, then migrate to Configurations. **Exactly the same evolution as Terraform users who start with inline resources and then extract modules.**

### Recommendation if building a platform for a company

- **Start with Configurations from day one** — the overhead is small, the benefits compound quickly
- **One git repo per Configuration** — treat it like a Terraform module repo with its own semantic versioning
- **Publish to your internal registry** — don't rely on the Upbound Marketplace for internal packages
- **Write a proper README** in `package/README.md` with a parameters table and "what gets deployed" section — the ecosystem won't generate it for you automatically (unlike the Terraform Registry)
- **Use GitOps for XR instances** — store XR YAML in a separate app-team repo; Flux/ArgoCD applies them

---

## Landing zone: one Configuration per capability

For a landing zone, model each capability as its own Configuration:

| Capability | Example Configuration | XR kind |
|------------|-----------------------|---------|
| Networking | `configuration-aws-network` | `XNetwork`, `XVPC` |
| Object storage | `configuration-aws-s3` | `XBucket`, `XStorageBucket` |
| Databases | `configuration-aws-rds` | `XDatabase`, `XPostgreSQL` |
| Compute / clusters | `configuration-aws-eks` | `XCluster`, `XEKS` |
| IAM / security | `configuration-aws-iam` | `XRole`, `XPolicy` |

Use `up project init` per capability, then build and push each as a separate package. See [Up CLI and DevEx](./17_Up_CLI_And_DevEx.md#landing-zone-one-configuration-per-capability) for the full workflow.

---

## Application compositions (v2) as Configurations

In v2, Compositions can compose **any** Kubernetes resource — not only cloud MRs. An **XApp** that creates a Deployment, Service, and ConfigMap can be packaged as a Configuration just like an XBucket.

| Component | Cloud (e.g. XBucket) | Application (e.g. XApp) |
|-----------|----------------------|--------------------------|
| XRD | `XBucket` | `XApp` |
| Composed resources | S3 Bucket (MR) | Deployment, Service, ConfigMap (core K8s) |
| Providers needed | provider-aws-s3 | None (Crossplane applies core resources) |
| Functions | function-patch-and-transform | function-patch-and-transform, function-auto-ready |
| RBAC | Provider handles it | ClusterRole aggregating to Crossplane |

**Configuration layout for XApp:**

```
configuration-xapp/
├── crossplane.yaml     # dependsOn: function-patch-and-transform, function-auto-ready
├── xrd.yaml            # XApp schema
├── composition.yaml    # pipeline → Deployment + Service + ConfigMap
└── rbac.yaml           # ClusterRole for deployments, services, configmaps
```

The package is identical in structure: XRD + Composition + optional RBAC. The only difference is that composed resources are core K8s resources instead of cloud MRs.

**Landing zone with apps:**

| Capability | Composed resources | Configuration |
|------------|-------------------|---------------|
| Storage | S3 Bucket (MR) | configuration-aws-s3 |
| Database | RDS Instance (MR) | configuration-aws-rds |
| **App** | **Deployment + Service + ConfigMap** | **configuration-xapp** |
| Full stack | RDS + Deployment + Service + Secret | configuration-fullstack |

See [examples/05-app-with-deployment](../examples/05-app-with-deployment/) and [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md).

---

## See also

- [Packages & CLI](./07_Packages_CLI.md) — building, pushing, and installing packages
- [Upbound Marketplace](./11_Upbound_Marketplace.md) — discovering published Configurations
- [Composition: XRDs, Compositions & Functions](./05_Composition_XRDs_Functions.md) — deep dive into the building blocks
- [examples/07-multi-resource-composition](../examples/07-multi-resource-composition/) — the example this guide is based on
- [dot-kubernetes on Upbound Marketplace](https://marketplace.upbound.io/configurations/devops-toolkit/dot-kubernetes/v2.0.21) — real-world Configuration reference
- [vfarcic/crossplane-kubernetes on GitHub](https://github.com/vfarcic/crossplane-kubernetes) — source repo for dot-kubernetes
