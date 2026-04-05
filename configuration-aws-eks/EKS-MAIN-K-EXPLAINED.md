# `functions/eks/main.k` explained (KCL + AWS resources)

This document walks through the **Upbound EKS composition function** written in **KCL**. It pairs with a **local snapshot** of the file:

- **Snapshot (copy):** [reference/eks-main.k](./reference/eks-main.k)  
- **Upstream (always current):** [github.com/upbound/configuration-aws-eks — `functions/eks/main.k`](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k)

If behavior differs from your cluster, **compare against upstream**; the snapshot can drift.

---

## 1. What this file is

The inputs **`oxr`** / **`_ocds`** and the output **`items`** are Crossplane’s **RunFunction** model in KCL form. See **[RUNFUNCTION-MODEL.md](./RUNFUNCTION-MODEL.md)** for the generic request/response picture.

| Concept | Meaning |
|--------|---------|
| **Composition Function** | A containerized step in a **Composition pipeline**. Crossplane passes **observed** XR + composed resources; the function returns **desired** resources as data. |
| **KCL (K Configuration Language)** | Declarative language used here to **build a list of Kubernetes-shaped objects** (`items`) from inputs. |
| **KRM KCL** | Pattern where those objects are **Crossplane managed resources** (and meta kinds like `CompositeConnectionDetails`). |

The function’s **output** is the list assigned to **`items = _items`** at the bottom: Crossplane applies that as **desired composed resources**.

### Is KCL “just generating YAML”?

**Almost** — the important part is **what** is generated and **how** it is delivered.

| True | Not quite |
|------|-----------|
| KCL **builds Kubernetes resource definitions** — the same *information* you would express in YAML manifests: `iamv1beta1.Role`, `eksv1beta1.Cluster`, `eksv1beta1.NodeGroup`, etc. | It is **not** “write a `.yaml` file to disk.” The function returns them through the **RunFunction** response; the apiserver **stores** them as API objects. |
| The result is **structured data** (typed against CRD models), which could be serialized as YAML for debugging. | It is **not** only dumb string templating: this file uses **logic** (`if`, `ready()`, `exists()`), **defaults**, and **merged specs** — so it is **programming** desired state, not only substituting a static template. |

**One line:** KCL here **generates the desired Crossplane / Kubernetes resource specs** — the same role as hand-written YAML — but **produced by code** and **returned to Crossplane**, not saved as a file. See **[RUNFUNCTION-MODEL.md](./RUNFUNCTION-MODEL.md)** for how that fits the request/response flow.

---

## 2. KCL building blocks used in this file

### Imports

```k
import models.io.upbound.awsm.eks.v1beta1 as eksv1beta1
```

Imports **typed schemas** for Crossplane CRDs (EKS, IAM, EC2, Kubernetes provider, Helm provider). You construct values like `eksv1beta1.Cluster { ... }` — KCL checks shapes against the model.

### Function inputs: `option("params")`

```k
oxr = option("params").oxr   # observed composite (your EKS XR)
_ocds = option("params").ocds # observed composed resources (by composition resource name)
_dxr = option("params").dxr   # desired composite (status patches, etc.)
```

- **`oxr`** — what the user applied (`spec.parameters`, `metadata`, …).  
- **`_ocds`** — what already exists from **previous** reconciles (used for **readiness** and **branching**).  
- **`_dxr`** — template for updating **composite status**; merged at the end.

### Optional chaining `?.`

```k
_ocds?.kubernetesCluster?.Resource?.status?.atProvider
```

Safe navigation: if any part is missing, the expression short-circuits instead of erroring (KCL rules apply).

### Lambdas (anonymous functions)

```k
_metadata = lambda name: str -> any {
    { annotations = { "krm.kcl.dev/composition-resource-name" = name } }
}
```

Defines a small reusable function. **`lambda arg: type -> returnType { body }`**.

### Dict merge `|`

```k
spec: _defaults | { forProvider = { region = params.region } }
```

Merges two dictionaries (right wins on key conflicts). Used to inject **`providerConfigRef`**, **`managementPolicies`**, and resource-specific **`spec`**.

### List append `+=`

```k
_items += [ resourceA, resourceB ]
```

Appends to the list of desired objects.

### Conditionals `if`

```k
if clusterSecurityGroupId:
    _items += [ ec2v1beta1.SecurityGroup { ... } ]
```

Only adds resources when the condition holds.

### `ready()` helper

Checks that an observed resource has **conditions** all **`True`** and that **`atProvider`** exists in **status** — i.e. provider has **successfully reconciled** enough for the next step.

### `exists()` helper

Returns whether a key exists in **`_ocds`** (composed resource already present from an earlier pass).

---

## 3. What gets created — summary table

| Composition name (`krm.kcl.dev/composition-resource-name`) | KCL type | Kubernetes / Crossplane kind (typical API) | AWS (conceptual) |
|------------------------------------------------------------|----------|--------------------------------------------|------------------|
| `controlplaneRole` | `iamv1beta1.Role` | IAM Role MR | **IAM role** for EKS control plane (`eks.amazonaws.com`) + `AmazonEKSClusterPolicy` |
| `kubernetesCluster` | `eksv1beta1.Cluster` | EKS Cluster MR | **EKS cluster** |
| `clusterSecurityGroupImport` | `ec2v1beta1.SecurityGroup` | EC2 SecurityGroup MR (import) | **VPC cluster security group** (observed, imported by ID) |
| `kubernetesClusterAuth` | `eksv1beta1.ClusterAuth` | ClusterAuth MR | EKS **auth data**; writes **kubeconfig** to Secret |
| `nodegroupRole` | `iamv1beta1.Role` | IAM Role MR | **IAM role** for worker nodes (`ec2.amazonaws.com`) + worker/CNI/EBS/ECR policies |
| `nodeGroupPublic` | `eksv1beta1.NodeGroup` | NodeGroup MR | **EKS managed node group** |
| *(optional)* access entry / policy | `AccessEntry`, `AccessPolicyAssociation` | EKS access MRs | **EKS access entry** + **cluster admin policy** for `iam.principalArn` |
| `vpc-cni-addon` | `eksv1beta1.Addon` | Addon MR | EKS **add-on** `vpc-cni` |
| `ebsCSIDriverRole` | `iamv1beta1.Role` | IAM Role MR | **IAM role** for EBS CSI via **EKS Pod Identity** |
| `ebsCSIDriverPodIdentityAssociation` | `eksv1beta1.PodIdentityAssociation` | PodIdentityAssociation MR | **Pod Identity association** for CSI SA |
| `aws-ebs-csi-driver-addon` | `eksv1beta1.Addon` | Addon MR | EKS **add-on** `aws-ebs-csi-driver` |
| `eks-pod-identity-agent-addon` | `eksv1beta1.Addon` | Addon MR | EKS **add-on** `eks-pod-identity-agent` |
| `providerConfig-kubernetes` | `k8sv1alpha1.ProviderConfig` | Kubernetes **ProviderConfig** | Not AWS — in-cluster **Kubernetes provider** creds |
| `providerConfig-helm` | `helmv1beta1.ProviderConfig` | Helm **ProviderConfig** | Not AWS — in-cluster **Helm provider** creds |
| *(meta)* | `CompositeConnectionDetails` | Connection details | Surfaces **kubeconfig** to XR connection |
| *(status)* | patched `_dxr` | XR **status** | `clusterName`, `nodeGroupArn` on composite |

**Not in this file:** VPC, subnets, Internet Gateway — those come from **`configuration-aws-network`**; this code only **selects subnets by label** matching `params.id`.

---

## 4. Execution order (why some resources appear “late”)

The KCL uses **`ready(...)`** and **`exists(...)`** so later resources wait for earlier ones (similar to Terraform dependencies, but expressed as **conditional list building**):

```
Always (first reconcile wave):
  controlplaneRole + kubernetesCluster
  nodegroupRole
  ebs CSIDriver role + PodIdentityAssociation (declared before addons that need readiness)

When cluster SG ID known:
  clusterSecurityGroupImport (EC2 SG)

When cluster ready OR ClusterAuth already exists:
  kubernetesClusterAuth → Secret with kubeconfig name {uid}-ekscluster

When vpc-cni addon ready OR node group already exists:
  nodeGroupPublic

When iam.principalArn set:
  AccessEntry + AccessPolicyAssociation

When ClusterAuth ready OR vpc-cni addon exists:
  vpc-cni-addon

When pod identity assoc + node group ready (or addon exists):
  aws-ebs-csi-driver-addon

When node group ready (or pod identity agent addon exists):
  eks-pod-identity-agent-addon

Always (late):
  Kubernetes + Helm ProviderConfigs (kubeconfig from Secret)
  CompositeConnectionDetails
  Merge status into _dxr and append to _items
```

So **one** `crossplane render` with **empty** `_ocds` often shows **only the first wave**; a live cluster fills `_ocds` over time. See [RENDER-LOCAL.md](./RENDER-LOCAL.md).

---

## 5. Section-by-section (maps to `reference/eks-main.k`)

### Defaults (`_defaults`)

Spreads into every AWS-tied MR:

- **`managementPolicies`** — usually `["*"]` (full lifecycle).  
- **`providerConfigRef`** — **`kind: ProviderConfig`**, **`name`** from **`params.providerConfigName`** (e.g. `default`). Matches why you need a **namespaced** `ProviderConfig` for this path.

### Control plane IAM + EKS cluster

- **`controlplaneRole`**: trust **`eks.amazonaws.com`**, attach **`AmazonEKSClusterPolicy`**.  
- **`kubernetesCluster`**: **`params.region`**, **`params.version`**, **`accessConfig`** from XR, **`roleArnSelector`** finds the role by label **`role: controlplane`**, **`vpcConfig`** uses **public subnets** with label **`networks.aws.platform.upbound.io/network-id` = `params.id`**.

**AWS:** IAM role + **EKS cluster** resource.

### Optional EC2 SecurityGroup import

If **`_ocds`** already has **`kubernetesCluster`** status with **`vpcConfig.clusterSecurityGroupId`**, creates an MR that **imports** that SG (external name = ID) for Crossplane management/discovery.

**AWS:** existing **security group** (cluster SG), not creating a new logical VPC resource here.

### ClusterAuth

After cluster is **ready** (or auth already exists): **`ClusterAuth`** writes connection secret **`{metadata.uid}-ekscluster`** in the XR namespace — this is what [HOW-IT-WORKS.md](./HOW-IT-WORKS.md) uses for **kubeconfig**.

### Node group IAM role

Worker role: **`ec2.amazonaws.com`**, policies for **worker node**, **VPC CNI**, **EBS CSI driver**, **ECR read**.

### Node group

Only when **`vpc-cni` addon is ready** or **`nodeGroupPublic` already exists** (bootstrap for upgrades):

- **`scalingConfig`**, **`instanceTypes`**, subnets same label selector as cluster (**public** + **`network-id`**).

**AWS:** **Managed node group** (EC2 instances ASG-style behind EKS).

### Optional IAM principal admin

If **`params.iam.principalArn`**: **`AccessEntry`** + **`AccessPolicyAssociation`** with **`AmazonEKSClusterAdminPolicy`**. Metadata name uses **SHA256** so changing ARN forces replace.

### Add-ons

- **`vpc-cni`** — after cluster auth ready or addon already tracked.  
- **`aws-ebs-csi-driver`** — after pod identity association **and** node group ready (or addon exists).  
- **`eks-pod-identity-agent`** — after node group ready (or agent addon exists).

**AWS:** EKS **add-on** API (managed add-on versions installed into the cluster).

### In-cluster ProviderConfigs

**Kubernetes** and **Helm** `ProviderConfig` objects use the **same kubeconfig Secret** so other compositions can target **this** EKS cluster from the management cluster.

### CompositeConnectionDetails + status

- Forwards **kubeconfig** connection detail when **`kubernetesClusterAuth`** is observed.  
- Patches **`status.eks.clusterName`** and **`nodeGroupArn`** from observed composed resources.

---

## 6. Related docs in this folder

| Doc | Role |
|-----|------|
| [CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md) | XR → Composition → Function → image → this file |
| [HOW-IT-WORKS.md](./HOW-IT-WORKS.md) | User-facing flow, kubeconfig, delete, `kubectl` |
| [RENDER-LOCAL.md](./RENDER-LOCAL.md) | `crossplane render` without cluster apply |
| [reference/eks-main.k](./reference/eks-main.k) | Byte snapshot of upstream `main.k` |

---

## 7. Further reading

- [KCL language](https://kcl-lang.io/docs/) — general language reference.  
- [Crossplane composition functions](https://docs.crossplane.io/latest/composition/compositions/) — how functions fit in the pipeline.  
- Upstream EKS function: [`functions/eks/main.k`](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k).
