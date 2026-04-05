# How to Discover XR Parameters (Terraform Variables Equivalent)

Crossplane doesn't have a `terraform plan` or `variable "x" {}` block you can read locally. The **XRD** (CompositeResourceDefinition) is the schema — it defines all parameters, required fields, defaults, and types. Here's how to find it and how it compares to Terraform.

---

## Terraform vs Crossplane: Where Variables Live

| Terraform | Crossplane |
|-----------|------------|
| `variable "region" { type = string }` in your module | XRD `spec.versions[].schema.openAPIV3Schema` |
| `terraform plan` shows what's needed | `kubectl get xrd <name> -o yaml` shows the schema |
| Module docs / `terraform-docs` | Package README, Marketplace, or XRD in cluster |
| `variables.tfvars` → `terraform apply` | `spec.parameters` in XR YAML → `kubectl apply` |

**The XRD is the source of truth.** Good Configuration packages document it; you can always inspect it after install.

---

## Curated EKS API vs full Terraform surface

A typical **Terraform** layout for EKS (VPC, subnets, tags, endpoints, encryption, addons, node groups per AZ, launch templates, etc.) can expose **many** `variable` blocks. The **`configuration-aws-eks`** **EKS** XR exposes a **smaller** set (`id`, `region`, `nodes`, `version`, `accessConfig`, `iam`, … — see the table below). That gap is **intentional**, not an accident.

| Fewer XR fields | Why it helps |
|-----------------|--------------|
| **Opinionated defaults** | The Composition / upstream KCL encodes VPC/Network wiring, IAM patterns, and node group shape; you do not re-specify every AWS knob on every cluster. |
| **Safer golden path** | Less surface for misconfigured networking or access for common internal platforms. |
| **Stable platform API** | Teams learn one **EKS** CR shape instead of an entire Terraform module’s variable list. |

**Trade-off:** Anything **not** on the XR needs another path: **fork** the Configuration, use **[eks-cluster-spot](../eks-cluster-spot/)** for provider-style control (e.g. Spot), or raw Terraform / MRs. Examples called out in this folder:

- **CNI / addons:** the stock package has **no** parameters for CNI choice or addon toggles — see **[CILIUM-TODO.md](./CILIUM-TODO.md)**.
- **Spot node groups:** no `capacityType` on the stock **EKS** API — see **`nodes`** in this file and [README — Configuration vs eks-cluster-spot](./README.md#configuration-vs-eks-cluster-spot-provider-based-eks).

**Summary:** A **narrow** **EKS** API matches Crossplane’s usual role as a **curated control-plane abstraction**; mirroring **every** Terraform variable would duplicate AWS surface area in the XRD. Fewer parameters than Terraform is usually **better for a packaged platform**; **full** flexibility stays in Terraform or forked Compositions when you need it.

**Broader pattern:** [Terraform modules vs Crossplane Configurations](../guides/20_Terraform_Modules_vs_Crossplane_Configurations.md) (why generic reusable modules map differently to Configurations).

---

## How to Discover Parameters

### 1. After installing the Configuration — inspect the XRD

```bash
# List XRDs installed by the Configuration
kubectl get xrd

# Get the full schema (parameters, types, required, defaults)
kubectl get xrd eks.aws.platform.upbound.io -o yaml
```

Look at `spec.versions[].schema.openAPIV3Schema.properties.spec.properties.parameters` — that's your "variables" block.

### 2. Package documentation

- **Marketplace:** [upbound/configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks) — often has parameter docs
- **README** in the package or upstream repo
- **Examples** in the package (e.g. `eks-xr-example.yaml`)

### 3. Upstream source

For configuration-aws-eks, the XRD lives in the repo:

- [apis/eks/definition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/definition.yaml)

---

## EKS XR Parameters (from XRD)

Full parameter reference for `kind: EKS` (apiVersion: `aws.platform.upbound.io/v1alpha1`):

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `id` | string | ✅ | — | ID of this cluster; other objects refer to it. Used for labels/selectors. |
| `region` | string | ✅ | — | AWS region (e.g. `eu-central-1`, `us-east-1`) |
| `providerConfigName` | string | ✅ | `default` | Crossplane ProviderConfig for AWS credentials |
| `nodes` | object | ✅ | — | Node group config (see below) |
| `version` | string | — | `"1.34"` | Kubernetes version. Enum: `1.31`, `1.32`, `1.33`, `1.34`, `1.35` |
| `accessConfig` | object | — | see below | Cluster access (auth mode, bootstrap permissions) |
| `iam` | object | — | — | IAM config (e.g. `principalArn` for cluster admin) |
| `managementPolicies` | array | — | `["*"]` | MR lifecycle: `*` or subset of Create, Observe, Update, Delete, LateInitialize |

### `nodes` (required)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `count` | integer | ✅ | — | Desired node count (**1–100** per XRD). `1` is valid (single worker; no HA). |
| `instanceType` | string | ✅ | `t3.small` | Instance type (e.g. `t3.small`, `m5.large`) |

**Spot:** the stock `EKS` API has **no** `capacityType` / Spot field. Upbound’s composition creates an on-demand managed node group. For Spot, fork the Configuration (set `capacityType: SPOT` on `NodeGroup` in KCL) or use provider-based **[eks-cluster-spot](../eks-cluster-spot/)** — see [README](./README.md#configuration-vs-eks-cluster-spot-provider-based-eks).

### `accessConfig` (optional)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `authenticationMode` | string | `CONFIG_MAP` | `CONFIG_MAP`, `API`, or `API_AND_CONFIG_MAP` |
| `bootstrapClusterCreatorAdminPermissions` | boolean | `true` | Bootstrap access config to cluster |

### `iam` (optional)

| Field | Type | Description |
|-------|------|-------------|
| `principalArn` | string | IAM principal ARN to connect as ClusterAdmin (creates access entry) |

---

## Minimal vs Full Example

**Minimal (required only):**

```yaml
apiVersion: aws.platform.upbound.io/v1alpha1
kind: EKS
metadata:
  name: my-cluster
  namespace: default
spec:
  parameters:
    id: my-cluster
    region: eu-central-1
    providerConfigName: default
    nodes:
      count: 2
      instanceType: t3.small
```

**Full (with optional fields):**

```yaml
apiVersion: aws.platform.upbound.io/v1alpha1
kind: EKS
metadata:
  name: my-cluster
  namespace: default
spec:
  parameters:
    id: my-cluster
    region: eu-central-1
    providerConfigName: default
    version: "1.34"
    accessConfig:
      authenticationMode: API_AND_CONFIG_MAP
      bootstrapClusterCreatorAdminPermissions: true
    iam:
      principalArn: arn:aws:iam::123456789012:user/my-admin
    nodes:
      count: 3
      instanceType: t3.medium
```

---

## General Workflow for Any Configuration

1. **Install** the Configuration: `kubectl apply -f configuration-aws-eks-install.yaml`
2. **Wait** until HEALTHY: `kubectl get configuration`
3. **Discover** parameters: `kubectl get xrd <xrd-name> -o yaml` or check package docs
4. **Create** XR YAML with `spec.parameters` matching the schema
5. **Apply** the XR: `kubectl apply -f my-eks-xr.yaml`

---

## See also

- [README.md — Curated API vs Terraform](./README.md#curated-api-vs-terraform) — short pointer; full discussion is this file, section **Curated EKS API vs full Terraform surface**
- [eks-xr-example.yaml](./eks-xr-example.yaml) — basic example
- [eks-xr-example-full.yaml](./eks-xr-example-full.yaml) — example with optional fields
- [HOW-IT-WORKS.md](./HOW-IT-WORKS.md) — how one XR becomes many AWS resources; **kubeconfig** (Step 4); **deleting the XR / async AWS teardown** (Step 5)
- [CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md) — **XR → XRD → Composition → Function → OCI image → `functions/eks/main.k`**
- [RUNFUNCTION-MODEL.md](./RUNFUNCTION-MODEL.md) — **RunFunction** request/response for composition functions
