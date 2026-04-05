# How the configuration-aws-eks Creates All EKS Resources

This document explains **in detail** how a single **EKS** XR (the YAML in `eks-xr-example.yaml`) ends up creating many AWS resources (IAM roles, EKS cluster, node group, addons, etc.). You only apply one resource; the rest is done by the **Configuration package** and its **Composition pipeline**.

---

## What you see in this folder vs what lives in the package

| In **this folder** (your repo) | In the **Configuration package** (installed from Marketplace) |
|--------------------------------|----------------------------------------------------------------|
| `configuration-aws-eks-install.yaml` – installs the package | The **XRD** (defines the `EKS` API), the **Composition** (pipeline), and the **embedded function** (the “recipe” in KCL code) |
| `eks-xr-example.yaml` – **one** EKS resource you apply | The logic that turns that one EKS spec into many composed resources (Cluster, NodeGroup, IAM, addons, …) |

So: **you only have the “front door” (install + one XR). The “recipe” that creates everything else is inside the package.** Below we show how that recipe is wired and what it creates.

---

## The big picture: from one XR to many resources

```
  YOU APPLY                          CROSSPLANE + CONFIGURATION PACKAGE
  ─────────                          ─────────────────────────────────

  eks-xr-example.yaml
  (one EKS XR with
   id, region, nodes…)    ──►  1. Crossplane sees the XR (kind: EKS).
                               2. It finds the Composition for EKS (from the installed Configuration).
                               3. The Composition runs a PIPELINE (two steps).
                               4. Step 1: Function "upbound-configuration-aws-ekseks" runs.
                                  → It reads your XR spec (parameters).
                                  → It runs KCL code (the "recipe") that outputs a LIST of
                                    composed resources: IAM Roles, Cluster, NodeGroup, Addons, etc.
                                  → Crossplane creates/updates those resources in the cluster.
                               5. Step 2: Function "crossplane-contrib-function-auto-ready" runs.
                                  → It marks the XR as Ready only when all composed resources are ready.
                               6. AWS providers (EKS, IAM, …) reconcile those composed resources
                                  → Real EKS cluster, node group, IAM roles, etc. are created in AWS.
```

So: **one XR in → Composition pipeline → one function outputs many “composed” resources → Crossplane and providers create them in Kubernetes and in AWS.**

---

## Step 1: What you apply (the only thing in this folder)

**eks-xr-example.yaml** is a single custom resource:

```yaml
apiVersion: aws.platform.upbound.io/v1alpha1
kind: EKS
metadata:
  name: tiho-eks-from-crossplane
  namespace: default
spec:
  parameters:
    id: prod-cluster-network
    region: eu-central-1
    providerConfigName: default
    version: "1.34"
    nodes:
      count: 2
      instanceType: t3.small
    # ... accessConfig, etc.
```

There is **no** Cluster, NodeGroup, or IAM YAML in this folder. This one object is the **input** to the Composition.

---

## Step 2: What the Configuration package installs

When you apply **configuration-aws-eks-install.yaml**, Crossplane installs the **Configuration** package. That package contains:

### A. The XRD (CompositeResourceDefinition)

Defines the **EKS** API: which fields exist under `spec.parameters` (id, region, version, nodes, accessConfig, iam, providerConfigName, etc.). That’s why your XR has `spec.parameters` and not random fields.

### B. The Composition (pipeline, not a list of resources)

The Composition does **not** list resources in YAML. It says: “when you see an EKS XR, run this **pipeline**”:

```yaml
# This lives inside the package; summarized here from the upstream repo.
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: eks.aws.platform.upbound.io
spec:
  compositeTypeRef:
    apiVersion: aws.platform.upbound.io/v1alpha1
    kind: EKS
  mode: Pipeline
  pipeline:
    - functionRef:
        name: upbound-configuration-aws-ekseks
      step: eks
    - functionRef:
        name: crossplane-contrib-function-auto-ready
      step: crossplane-contrib-function-auto-ready
```

So:

- **Step “eks”**: run the function **upbound-configuration-aws-ekseks** (the EKS “recipe”).
- **Step “crossplane-contrib-function-auto-ready”**: run the auto-ready function so the XR becomes Ready only when composed resources are ready.

The **actual list of resources** (Cluster, NodeGroup, IAM, etc.) is **not** in this Composition YAML. It is **inside the first function**, which is implemented in **KCL** (Kubernetes Configuration Language)—so the “recipe” is KCL code, not CUE or plain YAML (see below).

---

## Step 3: What the EKS function actually creates (the “recipe”)

The function **upbound-configuration-aws-ekseks** is an **embedded function**: its code is bundled in the Configuration package. Upbound’s source for it is here:

- **Repository:** [github.com/upbound/configuration-aws-eks](https://github.com/upbound/configuration-aws-eks)
- **Function code (KCL):** [functions/eks/main.k](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k)

That KCL code:

1. Reads the **observed XR** (`oxr`) – your `spec.parameters` (id, region, version, nodes, …).
2. Optionally reads **observed composed resources** (`_ocds`) – already-created resources, for ordering and readiness.
3. Builds a list **`_items`** of desired resources.
4. Returns that list as the **desired composed resources** for Crossplane to create/update.

Here is what that list contains (simplified from `main.k`). The function creates **at least**:

| Composed resource (logical name) | Kind (API) | What it is in AWS |
|----------------------------------|------------|--------------------|
| `controlplaneRole` | `iam...Role` | IAM role for the EKS control plane (trust `eks.amazonaws.com`, AmazonEKSClusterPolicy) |
| `kubernetesCluster` | `eks...Cluster` | The EKS cluster (version, region, role ref, VPC config from **network selector**) |
| `kubernetesClusterAuth` | `eks...ClusterAuth` | Writes kubeconfig into a Secret |
| `nodegroupRole` | `iam...Role` | IAM role for node group (trust `ec2.amazonaws.com`, EKS worker policies) |
| `nodeGroupPublic` | `eks...NodeGroup` | EKS managed node group (instance type, count from `params.nodes`, subnet selector) |
| `vpc-cni-addon` | `eks...Addon` | VPC CNI addon |
| `ebsCSIDriverRole` | `iam...Role` | IAM role for EBS CSI driver |
| `ebsCSIDriverPodIdentityAssociation` | `eks...PodIdentityAssociation` | Pod identity for EBS CSI |
| `aws-ebs-csi-driver-addon` | `eks...Addon` | EBS CSI driver addon |
| `eks-pod-identity-agent-addon` | `eks...Addon` | EKS Pod Identity Agent addon |
| ProviderConfig (Kubernetes + Helm) | `ProviderConfig` | So you can use the cluster (kubeconfig from Secret) |

It may also create:

- `clusterSecurityGroupImport` (SecurityGroup) – when the cluster’s security group ID is known.
- `accessEntry` + `accessPolicyAssociation` – when you set `iam.principalArn` (cluster admin).

Important: **VPC and subnets are not created by this function.** The Cluster and NodeGroup use **selectors** like:

```text
subnetIdSelector.matchLabels = {
  "networks.aws.platform.upbound.io/network-id" = params.id
  access = "public"
}
```

So they expect a **network** (VPC + subnets) that is labeled with `networks.aws.platform.upbound.io/network-id` = your `params.id` and `access = "public"`. That network typically comes from another Configuration: **configuration-aws-network**, which the configuration-aws-eks project lists as a dependency. When you install from the Marketplace, the full solution may include or require that network Configuration so that the EKS Composition can find subnets.

---

## Step 4: How to see what was created

After you apply the EKS XR and wait until it is Ready, you can **see** the composed resources and the underlying managed resources.

### See your XR and its status

```bash
kubectl get eks.aws.platform.upbound.io
kubectl describe eks.aws.platform.upbound.io tiho-eks-from-crossplane
```

In the status you may see things like `clusterName`, `nodeGroupArn`, and connection details (e.g. kubeconfig Secret name).

### Access the cluster from your laptop (kubeconfig)

When the EKS XR is **Ready**, you can use **`kubectl`** against the new cluster from your machine.

#### Option A — Kubeconfig Secret from Crossplane (matches this Configuration)

The composition creates **`ClusterAuth`**, which writes a kubeconfig into a **Secret** in the **same namespace as the EKS XR**. The Secret name is **`{metadata.uid}-ekscluster`** (the XR’s Kubernetes UID, not `metadata.name`).

```bash
NS=default   # namespace where your EKS XR lives (e.g. default)

# Find the secret (name ends with -ekscluster)
kubectl get secret -n "$NS" | grep ekscluster

# Set SECRET_NAME to the full name from the line above, e.g. abcd1234-...-ekscluster
SECRET_NAME=<paste-secret-name>

kubectl get secret "$SECRET_NAME" -n "$NS" -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/eks-crossplane.kubeconfig

export KUBECONFIG=/tmp/eks-crossplane.kubeconfig
kubectl get nodes
```

**Merge into your default `~/.kube/config`** (optional):

```bash
KUBECONFIG="$HOME/.kube/config:/tmp/eks-crossplane.kubeconfig" kubectl config view --flatten > ~/.kube/config.merged
mv ~/.kube/config.merged ~/.kube/config
kubectl config get-contexts
kubectl config use-context <context-name>
```

You still need **RBAC on the EKS cluster**: ensure your AWS IAM principal has access (e.g. `iam.principalArn` on the XR, or `bootstrapClusterCreatorAdminPermissions` for the creator, or map users in `aws-auth` if using that mode).

#### Option B — AWS CLI

If you have **`aws`** credentials and **`eks:DescribeCluster`** / **`aws eks update-kubeconfig`** rights:

```bash
aws eks update-kubeconfig --region <your-region> --name <cluster-name-in-aws>
kubectl get nodes
```

Get **`<cluster-name-in-aws>`** from the AWS console, from `kubectl describe` on the composed **Cluster** managed resource, or from the EKS XR **status** if it surfaces `clusterName`.

### See composed resources — list MRs and get the right API group

The **EKS XR** lives under `aws.platform.upbound.io`. Each composed object is a **managed resource (MR)** in the **same namespace** as the XR (e.g. `default`). MR **kind + API group** depend on which **EKS / IAM** provider you installed (Upbound **family** often uses **`*.aws.m.upbound.io`**; older lines use **`*.aws.upbound.io`**).

#### 1. Discover EKS / IAM kinds on your cluster

```bash
kubectl api-resources | grep -E 'eks\.aws|iam\.aws'
```

Examples you might see:

- `cluster`, `nodegroup`, `addon`, `clusterauth` → `eks.aws.m.upbound.io` (family) **or** `eks.aws.upbound.io`
- `roles` / `role` → `iam.aws.m.upbound.io` **or** `iam.aws.upbound.io`

Use the **exact** `NAME` + `APIVERSION` from this output in `kubectl get` / `describe`.

#### 2. List MRs in the XR namespace

Replace `NS` with your XR namespace (`default` in [eks-xr-example.yaml](./eks-xr-example.yaml)):

```bash
NS=default

kubectl get cluster.eks.aws.m.upbound.io -n "$NS"
kubectl get nodegroup.eks.aws.m.upbound.io -n "$NS"
kubectl get addon.eks.aws.m.upbound.io -n "$NS"
kubectl get clusterauth.eks.aws.m.upbound.io -n "$NS"
```

If those fail, swap **`eks.aws.m.upbound.io`** for **`eks.aws.upbound.io`** (from step 1).

**How you know MRs are there:** rows appear under **`NAME`** with **`SYNCED` / `READY`** columns. Names are usually **generated**, not the XR name alone, e.g.:

- `tiho-eks-from-crossplane-1a255ab7483f` (cluster)
- `tiho-eks-from-crossplane-c90838ba5913` (node group)

The prefix often matches your EKS XR **`metadata.name`** plus a suffix from the composition.

**Optional — all managed resources in the namespace** (if your Crossplane build supports it):

```bash
kubectl get managed -n "$NS"
```

#### 3. Inspect one MR (`describe` for Events / Conditions / finalizers)

Copy **`NAME`** from `kubectl get` (full string), then:

```bash
kubectl describe cluster.eks.aws.m.upbound.io tiho-eks-from-crossplane-1a255ab7483f -n default
kubectl describe nodegroup.eks.aws.m.upbound.io tiho-eks-from-crossplane-c90838ba5913 -n default
```

Use your **actual** names from `kubectl get`. **`describe`** shows **Events** (create/delete progress), **Status.Conditions**, and **finalizers** — essential when delete is slow or stuck.

#### 4. Optional: filter by composite label

Some clusters label MRs with the composite name (exact label key can vary by Crossplane version):

```bash
kubectl get nodegroup.eks.aws.m.upbound.io -n default -l 'crossplane.io/composite=tiho-eks-from-crossplane'
```

If nothing returns, rely on **`kubectl get … -n "$NS"`** and the name prefix.

### See the pipeline / function output (optional)

If your Crossplane version exposes pipeline state or events, you may see the “eks” step and the “crossplane-contrib-function-auto-ready” step. The important point: **the first step outputs the list of resources shown in the table above**; the second step only sets Ready when they are ready.

---

## Step 5: Deleting the EKS XR — commands, timing, AWS vs Kubernetes

### Optional: confirm the XR is healthy before you delete

```bash
kubectl get eks.aws.platform.upbound.io -n default
kubectl describe eks.aws.platform.upbound.io <xr-name> -n default
kubectl get nodes   # after kubeconfig (Step 4)
```

Look for **Ready / Synced** (or equivalent) conditions **True** on the XR and composed EKS MRs, and **nodes Ready** in the workload cluster.

### How to delete

Either form is equivalent:

```bash
kubectl delete -f eks-xr-example.yaml
# same as:
kubectl delete eks.aws.platform.upbound.io tiho-eks-from-crossplane -n default
```

Use the **namespace** where the XR lives (examples use `default`).

### “Deleted” in kubectl ≠ AWS is empty yet

When `kubectl` prints **`eks.aws.platform.upbound.io "…" deleted`**, the **composite (XR) object** is removed from the Kubernetes API (or delete has been accepted). **Tear-down in AWS is asynchronous.**

- Composed **managed resources** (`Cluster`, `NodeGroup`, `Addon`, IAM `Role`, etc.) may **still exist** for a while with **`READY False`** / **`SYNCED False`** while the provider deletes the real cloud resources and clears **finalizers**.
- The **AWS console** can still show the cluster **ACTIVE** for minutes; then **`DELETING`** for a long time while node groups, ENIs, and the control plane finish.

**Watch progress:**

```bash
kubectl get eks.aws.platform.upbound.io -n default
kubectl get cluster.eks.aws.m.upbound.io,nodegroup.eks.aws.m.upbound.io -n default
```

Adjust API groups if your cluster uses `eks.aws.upbound.io` instead of `eks.aws.m.upbound.io` (`kubectl api-resources | grep eks`).

**AWS CLI:**

```bash
aws eks describe-cluster --name <cluster-external-name> --region <region> --query 'cluster.status'
```

- **`DELETING`** → normal; wait (often **15–30+ minutes** for a full stack).
- **`ACTIVE`** long after you expected delete → inspect MR **Events** / **finalizers** (see below).

### Events during delete (e.g. NodeGroup)

Seeing **`DeletedExternalResource`** with **“Successfully requested deletion of external resource”** repeated many times (**`x11 over 5m`**, etc.) is **normal**: the controller **re-reconciles** until AWS finishes the delete and the object is gone.

### If delete seems stuck

Use the **full MR names** from `kubectl get nodegroup…` / `kubectl get cluster…` (see *See composed resources* above), then:

```bash
kubectl describe cluster.eks.aws.m.upbound.io <cluster-mr-name> -n default
kubectl describe nodegroup.eks.aws.m.upbound.io <nodegroup-mr-name> -n default
```

Example (your names will differ):

```bash
kubectl describe nodegroup.eks.aws.m.upbound.io tiho-eks-from-crossplane-c90838ba5913 -n default
```

Check **finalizers**, **Conditions**, and **Events** (e.g. dependency order, API errors, leaked load balancers / ENIs blocking subnet or cluster delete).

### Order for full teardown (EKS + Network)

1. Delete the **EKS** XR first and wait until composed **EKS** resources are gone (or clearly finishing).
2. Then delete the **Network** XR if you also want the VPC/subnets removed — avoids EKS still using subnets while the network disappears.

### Terraform `destroy` vs Crossplane delete (why it feels different)

**Wall-clock time** on the AWS side is mostly the **same**: both drive the **same EKS/AWS APIs**. Node group drain, instance termination, ENI cleanup, and cluster deletion are **slow** either way.

What differs is **UX**:

- **Terraform** `destroy` usually **blocks the terminal** until resources are gone (or errors).
- **Crossplane** often removes the **XR** from Kubernetes **quickly**; you then **watch** managed resources and **AWS** catch up. That can feel like “it deleted fast in kubectl but AWS is still running” — that mismatch is expected, not proof that Crossplane is slower than Terraform at the cloud layer.

---

## End-to-end flow (summary)

| Step | Who | What |
|------|-----|------|
| 1 | You | Apply `configuration-aws-eks-install.yaml` → Configuration package is installed (XRD + Composition + embedded function). |
| 2 | You | Apply `eks-xr-example.yaml` → One **EKS** XR exists with `spec.parameters`. |
| 3 | Crossplane | Sees XR of kind `EKS`; finds Composition for `EKS`; runs pipeline. |
| 4 | Function “eks” | Runs KCL code; reads XR `spec.parameters`; builds list of composed resources (IAM roles, Cluster, NodeGroup, Addons, ProviderConfigs, …); returns them. |
| 5 | Crossplane | Creates/updates those composed resources in the cluster. |
| 6 | Function “auto-ready” | Sets XR Ready when all composed resources are ready. |
| 7 | AWS providers | Reconcile the composed resources → real EKS cluster, node group, IAM roles, addons in AWS. |
| 8 | Network | VPC/subnets are expected to come from elsewhere (e.g. configuration-aws-network) with the right labels; Cluster and NodeGroup select them by label. |

So: **the “how” is: one EKS XR → one Composition pipeline → one KCL function that outputs many resources → Crossplane and providers create them.** The “recipe” is the KCL file in the upstream repo; this folder only has the install manifest and the one XR you apply.

---

## References

- **Layered pipeline (XR → Composition → Function → image → `main.k`):** [CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md)
- **KCL source explained (AWS mapping, phased logic):** [EKS-MAIN-K-EXPLAINED.md](./EKS-MAIN-K-EXPLAINED.md) + [reference/eks-main.k](./reference/eks-main.k)
- **RunFunction request/response:** [RUNFUNCTION-MODEL.md](./RUNFUNCTION-MODEL.md)
- **Configuration source (Composition + function):** [github.com/upbound/configuration-aws-eks](https://github.com/upbound/configuration-aws-eks)
- **The “recipe” (KCL):** [functions/eks/main.k](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k)
- **XRD (EKS API):** [apis/eks/definition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/definition.yaml)
- **Composition (pipeline):** [apis/eks/composition.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/composition.yaml)
- **Dependencies (network, providers, function-auto-ready):** [upbound.yaml](https://github.com/upbound/configuration-aws-eks/blob/main/upbound.yaml) in the repo
