# Upbound configuration-aws-eks (v2.0.0)

This folder contains manifests and documentation for the **Upbound AWS EKS Configuration**: creating EKS clusters with a **single custom resource** (Composite Resource) instead of separate provider manifests (Cluster, NodeGroup, IAM, VPC, etc.).

**Marketplace:** [upbound/configuration-aws-eks@v2.0.0](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks/v2.0.0)  
**Source:** [github.com/upbound/configuration-aws-eks](https://github.com/upbound/configuration-aws-eks)

---

## What’s in this folder

| File | Purpose |
|------|--------|
| **cluster-provider-config.yaml** | **Cluster-scoped AWS credentials:** `ClusterProviderConfig` named `default`. Apply **after** `provider-family-aws`. |
| **provider-config-default.yaml** | **Namespaced `ProviderConfig` named `default` in `default` namespace** — required for **configuration-aws-network**, which sets `providerConfigRef.kind: ProviderConfig` on composed EC2 MRs. Without it: `ProviderConfig.aws.m.upbound.io "default" not found`. Same Secret ref as cluster config. |
| **provider-family-aws.yaml** | **Required for Network:** Installs provider-family-aws (ec2.aws.m.upbound.io). See [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md) for "no matches for kind", ProviderConfig issues, and **why subnets/VPCs can look unnamed in AWS** (Name tags). |
| **configuration-aws-network-install.yaml** | Installs the Network Configuration (VPC + subnets). Required for EKS when using your own VPC. |
| **configuration-aws-eks-install.yaml** | Installs the EKS Configuration package (XRD, Composition, embedded function). |
| **kustomization.yaml** + **kustomize/** | **VPC + EKS with shared id:** Kustomize replacements keep Network and EKS in sync. See [VPC-AND-EKS-WITH-KUSTOMIZE.md](./VPC-AND-EKS-WITH-KUSTOMIZE.md). |
| **network-xr-example.yaml** | Standalone Network (VPC) XR — apply before EKS if using configuration-aws-network. |
| **eks-xr-example.yaml** | Example EKS XR — apply this to create one cluster after the Configuration is HEALTHY. |
| **eks-xr-example-full.yaml** | Full example with optional parameters (version, accessConfig, iam). |
| **XR-PARAMETERS.md** | **Parameter reference:** how to discover XR parameters (Terraform variables equivalent), full EKS schema table. |
| **VPC-AND-EKS-WITH-KUSTOMIZE.md** | **VPC + EKS workflow:** parameterize Network and EKS with Kustomize; single source of truth for `id`. |
| **README.md** | This file: overview, API, install/use, and comparison with provider-based EKS. |
| **HOW-IT-WORKS.md** | **Detailed explanation:** how one XR becomes many resources; Composition pipeline; what the function creates; **how to find MR API groups, list/describe composed resources** (`kubectl get` / `describe` with `eks.aws.m.upbound.io`); kubeconfig; deleting the XR (async AWS teardown); links to upstream KCL source. |
| **CONFIGURATION-PIPELINE.md** | **Systematic map:** XR → XRD → Composition → `Function` CR → OCI image (`configuration-aws-eks_eks`) → repo files (`functions/eks/main.k`, `composition.yaml`, `definition.yaml`); what is / isn’t in `main.k`. |
| **RENDER-LOCAL.md** | **Generate composed YAML without applying to a cluster:** `crossplane render` + `functions.yaml`, Docker, caveats for EKS readiness-gated KCL. |
| **EKS-MAIN-K-EXPLAINED.md** | **Line-by-line guide** to upstream `functions/eks/main.k`: KCL concepts, phased logic, table of composed resources → **AWS** mapping. |
| **RUNFUNCTION-MODEL.md** | **RunFunction request/response** (what Crossplane sends to composition functions, what they return; pipeline flow). |
| **reference/eks-main.k** | **Snapshot copy** of upstream `main.k` for offline reading (prefer [upstream](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k) for truth). |
| — | **[AWS family: contrib vs Upbound](../guides/18_AWS_Family_Contrib_vs_Upbound.md)** — do not install two `provider-family-aws` lines; errors like `already controlled by ProviderRevision`. |

---

## Can you use it to create an EKS cluster?

**Yes.** Once the Configuration is installed, you create one **EKS** XR (e.g. from **eks-xr-example.yaml**). The configuration’s Composition and embedded function provision the underlying AWS resources (VPC, subnets, IAM, EKS cluster, node group, etc.) for you.

---

## How it works

| Component | Role |
|-----------|------|
| **XRD (CompositeResourceDefinition)** | Defines the **EKS** API: high-level schema (id, region, version, nodes, IAM, accessConfig, providerConfigName). |
| **Composition** | Pipeline mode: runs an **embedded function** that turns that spec into concrete managed resources. |
| **Embedded function** | KCL code that outputs the list of composed resources (IAM roles, Cluster, NodeGroup, addons, etc.); VPC/subnets come from a network Configuration (e.g. configuration-aws-network) selected by labels. |

You don’t create `Cluster`, `NodeGroup`, or IAM manifests yourself; one **EKS** XR composes everything. For a **step-by-step explanation** of how one XR becomes many resources, how to inspect what was created, **kubeconfig for local `kubectl`**, and **what happens when you delete the XR** (kubectl “deleted” vs AWS still cleaning up), see **[HOW-IT-WORKS.md](./HOW-IT-WORKS.md)** (*Step 4* — access; *Step 5* — deletion).

For a **compact layered view** (XR → Composition → Function → image → Git files), see **[CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md)**.

### Curated API vs Terraform

The **EKS** XR has **fewer** inputs than a large **Terraform EKS + VPC** module — that is **by design** (opinionated defaults, smaller platform API). What is not exposed on the XR (CNI choice, Spot, exotic addons, …) requires a fork, **[eks-cluster-spot](../eks-cluster-spot/)**, or raw Terraform — see **[XR-PARAMETERS.md — Curated EKS API vs full Terraform surface](./XR-PARAMETERS.md#curated-eks-api-vs-full-terraform-surface)**.

---

## The custom API (XRD)

- **API group:** `aws.platform.upbound.io`
- **Kind:** `EKS`
- **Version:** `v1alpha1`

**Main parameters:** `id`, `region`, `providerConfigName`, `version`, `nodes` (count, instanceType), `accessConfig`, `iam` (e.g. principalArn). See **[XR-PARAMETERS.md](./XR-PARAMETERS.md)** for the full parameter table and how to discover parameters (Terraform variables equivalent).

---

## Install and use

### Before everything works (VPC + Network + EKS)

Use this **ordered checklist** so you don’t hit the usual gaps. Skipping a step often shows up only when the Network or EKS XR starts reconciling.

| Step | What to do | Why it matters |
|------|------------|----------------|
| **1** | **Crossplane** 2.0+ (or UXP 2.0+) installed | Nothing reconciles without it. |
| **2** | **One** AWS family provider only — [provider-family-aws.yaml](./provider-family-aws.yaml) (Upbound) **or** contrib, not both | Two families fight over the same APIs; stale `ProviderRevision` errors. See [guide 18: contrib vs Upbound](../guides/18_AWS_Family_Contrib_vs_Upbound.md). |
| **3** | **Secret** `aws-secret` in `crossplane-system`, key `creds` (standard AWS credentials file content) | Both provider config kinds read this. |
| **4** | Apply **provider-family-aws**, then wait until **HEALTHY** (`kubectl get providers.pkg.crossplane.io`) | **configuration-aws-network** composes `ec2.aws.m.upbound.io` resources; those CRDs come from the **family** package. Without this: *no matches for kind `Subnet` / `VPC` in `ec2.aws.m.upbound.io`*. Details: [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md). |
| **5** | Verify CRDs exist, e.g. `kubectl get crd clusterproviderconfigs.aws.m.upbound.io` and `kubectl get crd subnets.ec2.aws.m.upbound.io` | **`ClusterProviderConfig`** CRD is not in Crossplane core — it arrives with the family. Applying cluster config *before* CRDs exist fails. |
| **6** | Apply **[cluster-provider-config.yaml](./cluster-provider-config.yaml)** and **[provider-config-default.yaml](./provider-config-default.yaml)** | **Two different objects:** `ClusterProviderConfig` is cluster-wide; **Network** composed MRs use **`ProviderConfig`** (namespaced) with `name: default`. **Only** cluster config → `ProviderConfig.aws.m.upbound.io "default" not found`. The namespaced config must live in the **same namespace** as the Network XR (examples use `default`). |
| **7** | Apply **configuration-aws-network-install.yaml**, wait until that **Configuration** is **HEALTHY** | Defines Network XRD/Composition; pulls EC2 provider via `dependsOn`. |
| **8** | Apply **Network** XR (`network-xr-example.yaml` or [kustomize](./VPC-AND-EKS-WITH-KUSTOMIZE.md)), wait until VPC/subnets are Ready | EKS expects subnets labeled with your `id`; create the network first. |
| **9** | Apply **configuration-aws-eks-install.yaml**, wait **HEALTHY** | EKS XRD, Composition, function. |
| **10** | Apply **EKS** XR; **`spec.parameters.id` must match** the Network’s `spec.parameters.id` | Same label `networks.aws.platform.upbound.io/network-id` on subnets; mismatch → EKS never finds subnets. Kustomize flow: [VPC-AND-EKS-WITH-KUSTOMIZE.md](./VPC-AND-EKS-WITH-KUSTOMIZE.md). |

**One-shot (this repo):** after step 3, `kubectl apply -k configuration-aws-eks/` from the repo root applies family, both provider configs, both Configurations, and the kustomized Network + EKS XRs — still ensure providers/Configurations become HEALTHY before expecting AWS resources.

**Symptoms ↔ what was missing**

| What you see | Usually means |
|--------------|----------------|
| `no matches for kind … ec2.aws.m.upbound.io` | Step **4** — install [provider-family-aws](./provider-family-aws.yaml), wait for CRDs. |
| `ProviderConfig … "default" not found` | Step **6** — add [provider-config-default.yaml](./provider-config-default.yaml) in the **Network XR’s namespace** (not only `ClusterProviderConfig`). |
| `no matches for kind ClusterProviderConfig` | Step **5** — apply cluster config **after** family CRDs exist. |
| EKS stuck / no subnets | Step **10** — **id** mismatch between Network and EKS, or Network not Ready yet. |
| `ProviderRevision` / “already controlled” | Step **2** — remove duplicate family; see [guide 18](../guides/18_AWS_Family_Contrib_vs_Upbound.md). |

---

**Prerequisites (short):**

| Requirement | Details |
|-------------|---------|
| Crossplane 2.0+ or UXP 2.0+ | Core Crossplane installed |
| **Providers** | configuration-aws-eks auto-installs **provider-aws-eks** (and deps) via `dependsOn`. For **Network** (VPC): apply [provider-family-aws.yaml](./provider-family-aws.yaml) for the AWS **family** layer; **provider-aws-ec2** is then installed by **configuration-aws-network**’s `dependsOn` — so you may only apply the family manifest yourself, and EC2 still appears as a separate `Provider` row. See comments in [provider-family-aws.yaml](./provider-family-aws.yaml). [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md) for API-group issues. |
| **AWS credentials** | **Both** [cluster-provider-config.yaml](./cluster-provider-config.yaml) (`ClusterProviderConfig` `default`) **and** [provider-config-default.yaml](./provider-config-default.yaml) (`ProviderConfig` `default` in namespace `default`) for this flow: EKS/other paths may use cluster config; **Network** compositions require the namespaced `ProviderConfig` in the XR’s namespace. Secret `aws-secret` (key `creds`) in `crossplane-system`. Apply after `provider-family-aws` and CRDs exist. |

**EKS-only quick path** (no configuration-aws-network VPC — uses defaults inside the function or another network story):

```bash
# From this folder
kubectl apply -f configuration-aws-eks-install.yaml
kubectl get configuration   # wait until configuration-aws-eks is HEALTHY
kubectl apply -f eks-xr-example.yaml
kubectl get eks.aws.platform.upbound.io   # wait until READY
```

Alternatively: use the **Install Manifest** on the [Marketplace page](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks/v2.0.0), or run **`up project run`** from the [configuration repo](https://github.com/upbound/configuration-aws-eks).

---

## Who creates ClusterRole, RoleBinding, and ServiceAccount?

| Package type | Who installs Kubernetes RBAC? | What gets created |
|--------------|--------------------------------|-------------------|
| **Provider** (`pkg.crossplane.io/v1` `Provider`) | **Crossplane’s package controller** reads RBAC manifests **embedded in the provider OCI image** and applies them to the cluster. | `ClusterRole` / `ClusterRoleBinding` (or `Role` / `RoleBinding`) for the provider’s **ServiceAccount** (e.g. `upbound-provider-aws-ec2-*` in `crossplane-system`). The provider controller needs permission to **list/watch** `ProviderConfig`, `ClusterProviderConfig`, and all **MR** CRDs it reconciles (e.g. `providerconfigs.aws.m.upbound.io`). |
| **Configuration** (`pkg.crossplane.io/v1` `Configuration`) | Crossplane installs **XRDs, Compositions, Functions** from the package. **It does not** ship the AWS (or other cloud) provider controller’s RBAC — that comes only from **Provider** packages. | No provider `ClusterRole` for EC2/EKS from a Configuration alone. |
| **Crossplane core** (Helm / install manifest) | Your **Crossplane install** (e.g. Helm chart) creates RBAC for the Crossplane and RBAC manager pods. | Separate from per-provider RBAC. |

**You do not** hand-write `ClusterRoleBinding` for official providers under normal operation. If you see errors like `cannot list resource "providerconfigs" ... aws.m.upbound.io is forbidden`, the provider’s ServiceAccount is missing rules — usually fixed by **reinstalling** or **upgrading** the Provider so the package manager reapplies RBAC, or by aligning **provider-family-aws** + **provider-aws-ec2** versions.

**Inspect provider RBAC:**

```bash
kubectl get clusterrolebinding | grep -E 'provider|aws'
kubectl get clusterrole | grep -E 'provider|aws'
kubectl get sa -n crossplane-system | grep provider
```

More context: [Configurations guide — Provider packages and RBAC](../guides/13_Configurations.md#provider-packages-and-kubernetes-rbac).

---

## Configuration vs eks-cluster-spot (provider-based EKS)

| Aspect | **configuration-aws-eks** (this folder) | **eks-cluster-spot/** (provider-based) |
|--------|------------------------------------------|----------------------------------------|
| **Model** | One XR → Composition → many resources | Separate manifests: Providers, IAM, Cluster, NodeGroup, VPC, subnets. |
| **API** | Custom `EKS` (aws.platform.upbound.io) | Raw provider CRs: Cluster, NodeGroup, Role, etc. |
| **Flexibility** | Tuned to the schema; less low-level control unless you fork. | Full control: Spot, custom subnets, multiple node groups. |
| **VPC/subnets** | Typically created by the embedded function. | provider-ec2 (vpc.yaml, subnets.yaml) or bring your own. |
| **Spot instances** | Depends on function support; not in the basic example. | Explicit: `capacityType: SPOT` in nodegroup-spot.yaml. |

**Use this folder** when you want a simple “one resource” EKS API.  
**Use [eks-cluster-spot/](../eks-cluster-spot/)** when you need Spot node groups, specific VPC/subnet layout, or full control over each provider resource.
