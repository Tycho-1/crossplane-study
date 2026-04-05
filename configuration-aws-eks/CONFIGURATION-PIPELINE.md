# configuration-aws-eks: XR → Composition → Function → Image → Source

Systematic view of how the **Upbound AWS EKS Configuration** goes from the YAML you apply to the **OCI image** that runs **KCL**, and which **repo files** correspond to each layer.

**Upstream repo:** [github.com/upbound/configuration-aws-eks](https://github.com/upbound/configuration-aws-eks)  
**This study folder:** install manifests + examples only; the pipeline below lives in the **package** installed from the Marketplace.

---

## 1. End-to-end chain (order matters)

| Step | What it is | Where it lives |
|------|------------|----------------|
| **1** | **XR (Composite Resource)** — your `kind: EKS` instance | **Your cluster** — e.g. [eks-xr-example.yaml](./eks-xr-example.yaml) |
| **2** | **XRD** — defines the `EKS` API (`spec.parameters`, schema) | **Configuration package** — [apis/eks/definition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/definition.yaml) |
| **3** | **Composition** — pipeline: which **Functions** run, in order | **Configuration package** — [apis/eks/composition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/composition.yaml) |
| **4** | **Function** (`pkg.crossplane.io/v1`) — installs & runs the function **runtime** | **Your cluster** — e.g. `Function` named `upbound-configuration-aws-ekseks` |
| **5** | **OCI image / xpkg** — container + metadata Crossplane executes | **Registry** — e.g. `xpkg.upbound.io/upbound/configuration-aws-eks_eks@sha256:…` |
| **6** | **Source code** — **KCL** that **emits** composed managed resources | **Git repo** — [functions/eks/main.k](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k) (built into the image in step 5) |

**One sentence:** You apply an **EKS XR** → Crossplane uses the **XRD** + **Composition** → the **Composition** calls a **Function** whose **image** was built from **`functions/eks/main.k`** → that code outputs **desired MRs** (Cluster, NodeGroup, IAM, …) → **AWS providers** reconcile them to AWS.

---

## 2. Layer 1 — Composite Resource (XR)

- **Kind:** `EKS`  
- **API:** `aws.platform.upbound.io/v1alpha1`  
- **You provide:** `spec.parameters` (`id`, `region`, `nodes`, `version`, `accessConfig`, …).

Example: [eks-xr-example.yaml](./eks-xr-example.yaml).

The XR is only **desired state** for “one logical EKS stack.” It does **not** contain the full list of AWS objects — that list is **produced later** by the function.

---

## 3. Layer 2 — CompositeResourceDefinition (XRD)

- **Links** the composite kind (`EKS`) to a **Composition** (by name/labels, depending on setup).  
- **OpenAPI schema** for `spec.parameters` (what fields are valid on the XR).

**File in repo:** [apis/eks/definition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/definition.yaml)

---

## 4. Layer 3 — Composition (pipeline mode)

- **`mode: Pipeline`** — steps run in order.  
- Each step references a **`Function`** by **`functionRef.name`** (must match the `Function` object installed on the cluster).

**Upstream pipeline** ([apis/eks/composition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/composition.yaml)):

| Step | `functionRef.name` | Role |
|------|---------------------|------|
| `eks` | `upbound-configuration-aws-ekseks` | Runs **KCL**; outputs composed resources (IAM, EKS, addons, …). |
| `crossplane-contrib-function-auto-ready` | `crossplane-contrib-function-auto-ready` | Marks the XR **Ready** when composed resources are ready (no extra AWS MRs). |

The **Composition YAML does not** usually list every AWS field. It **delegates** the “recipe” to the **eks** function.

---

## 5. Layer 4 — `Function` resource on the cluster

When you install the Configuration, Crossplane also reconciles **`Function`** objects (from the package). Example (from your cluster):

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: upbound-configuration-aws-ekseks
spec:
  package: xpkg.upbound.io/upbound/configuration-aws-eks_eks@sha256:…
```

- **`metadata.name`** must match **`functionRef.name`** in the Composition.  
- **`spec.package`** is the **OCI reference** for this function’s **image** (see next section).

---

## 6. Layer 5 — OCI image (`configuration-aws-eks_eks`)

- The package name **`configuration-aws-eks_eks`** is the **function sub-package** built from the **`functions/eks/`** tree.  
- **Crossplane** pulls the image, runs the function **controller / runtime**, and when the pipeline invokes this step, your **`main.k`** logic runs and returns **desired composed resources**.

You do **not** edit the image on the cluster; you change **source** and **rebuild / republish** (Upbound’s CI does this for official releases).

---

## 7. Layer 6 — Repository files that define behavior

| Path (in [configuration-aws-eks](https://github.com/upbound/configuration-aws-eks)) | Purpose |
|-------------------------------------------------------------------------------------|--------|
| **`functions/eks/main.k`** | **KCL** — builds `_items`: IAM roles, `Cluster`, `NodeGroup`, `Addon`, `ClusterAuth`, optional `AccessEntry`, EBS CSI / pod identity pieces, in-cluster `ProviderConfig`s, etc. This is what **creates the list of MRs** for the EKS stack **inside this Configuration**. **Explained:** [EKS-MAIN-K-EXPLAINED.md](./EKS-MAIN-K-EXPLAINED.md); **local snapshot:** [reference/eks-main.k](./reference/eks-main.k). |
| **`functions/eks/kcl.mod`** | KCL module metadata (imports, deps). |
| **`apis/eks/composition.yaml`** | Pipeline: **which** function runs first (`eks`), then **auto-ready**. |
| **`apis/eks/definition.yaml`** | **XRD** — `EKS` API schema. |
| **`upbound.yaml`** | **Project** metadata: dependencies (e.g. `configuration-aws-network`, `provider-aws-eks`, `provider-aws-iam`, `function-auto-ready`, …). Does not replace `main.k`; it declares what the overall project expects. |

**Build / packaging:** Upbound uses **`up project build`** (and CI) to produce **multiple** artifacts from one repo — the **main Configuration** image (XRD + Composition + references) and the **embedded function** image **`configuration-aws-eks_eks`** that contains the compiled KCL bundle. Exact CI is in the repo’s `.github` workflows; locally, developers use the **`up`** CLI per [Upbound docs](https://docs.upbound.io/).

---

## 8. What is *not* defined in `main.k` (important)

| Topic | Where it comes from |
|--------|---------------------|
| **VPC, subnets, IGW, routing** | **`configuration-aws-network`** — different Configuration; [functions/network/main.k](https://github.com/upbound/configuration-aws-network/blob/main/functions/network/main.k) (not the EKS repo). |
| **Raw AWS resources you add yourself** | Your own Compositions / MRs / other packages. |

The EKS **`main.k`** **selects subnets by label** (`networks.aws.platform.upbound.io/network-id` = `spec.parameters.id`); it does **not** create those subnets.

---

## 9. Diagram (text)

```
  eks-xr-example.yaml  (XR: kind EKS, spec.parameters)
            │
            ▼
  XRD (definition.yaml)  ──►  validates / binds API
            │
            ▼
  Composition (composition.yaml)  ──►  pipeline step "eks"
            │
            ▼
  Function CR  upbound-configuration-aws-ekseks
            │
            ▼
  OCI package  configuration-aws-eks_eks@sha256:…
            │
            ▼
  Runtime executes KCL in  functions/eks/main.k
            │
            ▼
  Desired MRs  (Cluster, NodeGroup, Role, Addon, …)
            │
            ▼
  provider-aws-eks / provider-aws-iam / …  ──►  AWS API
```

---

## 10. See also

- **[RUNFUNCTION-MODEL.md](./RUNFUNCTION-MODEL.md)** — Crossplane **RunFunction** request/response (same contract for KCL, CUE, Go, …).
- **[RENDER-LOCAL.md](./RENDER-LOCAL.md)** — use **`crossplane render`** to print composed YAML **without** `kubectl apply` (still uses Docker to run functions).
- **[HOW-IT-WORKS.md](./HOW-IT-WORKS.md)** — narrative walkthrough, kubeconfig, delete behavior, listing MRs.  
- **[XR-PARAMETERS.md](./XR-PARAMETERS.md)** — `spec.parameters` table.  
- **Marketplace:** [upbound/configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks)
