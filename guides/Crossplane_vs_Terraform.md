# Crossplane vs Terraform: Comparison & Where to Find “Resource Docs”

This guide compares **Crossplane** and **Terraform** for creating infrastructure and explains **where to find the Crossplane equivalent** of Terraform’s AWS provider docs (e.g. [Terraform AWS Provider – Resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster)).

---

## Quick comparison

| Aspect | Terraform | Crossplane |
|--------|-----------|------------|
| **What it is** | CLI tool + state file; provisions infra via provider plugins | Control plane running in Kubernetes; provisions infra via CRDs + providers |
| **Language** | HCL (`.tf`) | YAML (Kubernetes manifests) |
| **API** | `terraform apply`, plan, state | `kubectl apply`, Kubernetes API, GitOps |
| **State** | Central state file (local or remote), locking | No separate state file; desired state = CR spec; cluster is source of truth |
| **“Resource docs”** | One place: [Terraform Registry – AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (resources, data sources, arguments) | Several places: Upbound Marketplace, doc.crds.dev, GitHub CRDs, `kubectl explain` (see below) |
| **Abstraction** | Modules | Compositions + XRDs (custom APIs) |
| **Self-service** | Shared modules, workspaces | Kubernetes RBAC + custom XRs per namespace/team |
| **Best for** | Teams without K8s, multi-cloud CLI workflows, classic IaC | K8s-native platforms, GitOps, self-service infra inside K8s |

### Do Crossplane and Terraform replace each other?

**Not by default.** Terraform remains the right tool for many **bootstrap** and **non-K8s** flows; Crossplane fits **in-cluster** control planes and **CR-based** self-service. They often **coexist**.

**Helm** is separate again: chart packaging and release semantics on Kubernetes. Crossplane does **not** obsolete Helm across the board — you may run **both** (e.g. `provider-helm` for some charts, standalone Helm elsewhere). For a fuller conclusion — **Helm + Terraform + Crossplane as complementary layers**, partial substitution only where you choose it — see **[Terraform modules vs Crossplane Configurations](./20_Terraform_Modules_vs_Crossplane_Configurations.md#helm-terraform-and-crossplane-complement-not-wholesale-replacement)**.

---

## Where Terraform puts “what can be done on AWS”

For Terraform you have **one main entry point**:

- **AWS Provider docs:**  
  [https://registry.terraform.io/providers/hashicorp/aws/latest/docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
  - **Resources:** [https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources)  
    Example: [eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster) lists all arguments, attributes, and examples.
  - **Data sources:** separate section for read-only lookups.

So: **one provider, one docs site, one list of resources and options.**

---

## Where Crossplane puts “what can be done on AWS”

Crossplane doesn’t have a single “AWS provider docs” page like Terraform. You use a mix of:

### 1. **Upbound Marketplace (main “registry” for Upbound AWS)**

- **Browse providers:**  
  [https://marketplace.upbound.io/providers](https://marketplace.upbound.io/providers)
- **AWS “family” (overview):**  
  [https://marketplace.upbound.io/providers/upbound/provider-family-aws/latest](https://marketplace.upbound.io/providers/upbound/provider-family-aws/latest)  
  Points to API reference and managed resources.
- **Per-service packages (closest to “resource groups”):**  
  AWS is split into **multiple packages** (e.g. EKS, S3, EC2). Each package has its own page and **Managed Resources** list:
  - **EKS (18 managed resources):**  
    [https://marketplace.upbound.io/providers/upbound/provider-aws-eks/latest](https://marketplace.upbound.io/providers/upbound/provider-aws-eks/latest) → **Managed Resources** section.
  - **S3:**  
    [https://marketplace.upbound.io/providers/upbound/provider-aws-s3/latest](https://marketplace.upbound.io/providers/upbound/provider-aws-s3/latest) → Managed Resources.
  - **EC2, RDS, IAM, etc.:** search or browse [Providers](https://marketplace.upbound.io/providers) for `provider-aws-*`.

So: **“Same info as Terraform’s AWS resource list”** = **Upbound Marketplace → choose provider (e.g. provider-aws-eks) → Managed Resources** (and linked API reference).

### 2. **doc.crds.dev (CRD reference, schema only)**

- **Auto-generated CRD docs** (fields, types, no prose like Terraform’s docs):
  - **Older Crossplane AWS provider:**  
    [https://doc.crds.dev/github.com/crossplane/provider-aws](https://doc.crds.dev/github.com/crossplane/provider-aws) (many CRDs).
  - **Upbound provider-aws:**  
    [https://doc.crds.dev/github.com/upbound/provider-aws](https://doc.crds.dev/github.com/upbound/provider-aws) (versioned; some packages may show under the main repo or family).

Use this for **exact field names and types** (like “what arguments exist for this resource”).

### 3. **Source: GitHub CRDs**

- **Upjet-based AWS provider (what Upbound builds from):**  
  [https://github.com/crossplane-contrib/provider-upjet-aws](https://github.com/crossplane-contrib/provider-upjet-aws)  
  CRDs are in the repo (e.g. per-package or in `package/crds`). Useful if you want to see the raw CRD YAML.

### 4. **In-cluster: `kubectl explain`**

After installing a provider, you get the CRDs in the cluster. Then:

```bash
# List CRDs installed by a provider (e.g. EKS)
kubectl get crd | grep eks

# Explain a specific resource (field names, types, required)
kubectl explain cluster.eks.aws.upbound.io
kubectl explain cluster.eks.aws.upbound.io.spec
kubectl explain cluster.eks.aws.upbound.io.spec.forProvider
```

This is the **cluster-local equivalent** of “what options does this resource support?”

---

## Mapping: Terraform resource → Crossplane resource

| Terraform (AWS) | Crossplane / Upbound equivalent | Where to look |
|-----------------|----------------------------------|----------------|
| [aws_eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster) | EKS **Cluster** MR (e.g. `eks.aws.upbound.io` or `eks.aws.m.upbound.io`) | Marketplace [provider-aws-eks](https://marketplace.upbound.io/providers/upbound/provider-aws-eks/latest) → Managed Resources; `kubectl explain cluster.eks...` |
| [aws_s3_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket) | S3 **Bucket** MR | Marketplace [provider-aws-s3](https://marketplace.upbound.io/providers/upbound/provider-aws-s3/latest); `kubectl explain bucket.s3...` |
| [aws_db_instance](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance) | RDS **Instance** MR | Marketplace provider-aws (or RDS sub-package); `kubectl explain instance.rds...` |
| Other `aws_*` resources | Corresponding MR in a `provider-aws-*` package | Marketplace → Providers → filter AWS; or doc.crds.dev; or `kubectl get crd \| grep aws` |

Naming may differ (e.g. `Cluster` vs `eks_cluster`). The **spec** of the MR (often `spec.forProvider`) holds the same kind of options as Terraform’s resource arguments.

---

## Side-by-side: EKS cluster example

**Terraform** ([eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster)):

```hcl
resource "aws_eks_cluster" "example" {
  name     = "my-eks"
  role_arn = aws_iam_role.eks.arn
  vpc_config {
    subnet_ids = var.subnet_ids
  }
  version = "1.28"
  # ... many more arguments in the docs
}
```

**Crossplane** (same “what can be done” – options come from Marketplace / CRD / `kubectl explain`):

```yaml
apiVersion: eks.aws.upbound.io/v1beta1
kind: Cluster
metadata:
  name: my-eks
spec:
  forProvider:
    name: my-eks
    roleArn: arn:aws:iam::123456789012:role/eks-role
    version: "1.28"
    vpcConfig:
      subnetIds:
        - subnet-xxx
        - subnet-yyy
  # ... same kind of options as Terraform, under forProvider
```

So: **Terraform’s “big doc on what can be used”** for a resource = **Crossplane’s CRD schema** (Marketplace Managed Resources + doc.crds.dev + `kubectl explain`).

---

## Summary: where to find “the same info as Terraform’s AWS docs”

1. **Upbound Marketplace**  
   [https://marketplace.upbound.io/providers](https://marketplace.upbound.io/providers)  
   → pick **provider-aws-eks**, **provider-aws-s3**, etc.  
   → open **Managed Resources** (and linked API reference).  
   This is the **closest equivalent** to Terraform’s [AWS provider resources list](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources).

2. **doc.crds.dev**  
   For **exact field names and types** (like Terraform’s “Arguments Reference”):  
   [https://doc.crds.dev/github.com/crossplane/provider-aws](https://doc.crds.dev/github.com/crossplane/provider-aws) (and Upbound provider if listed).

3. **In your cluster**  
   After installing the provider:  
   `kubectl get crd | grep <service>` and `kubectl explain <resource>.<group>`.

4. **Crossplane core concepts**  
   [https://docs.crossplane.io/latest/managed-resources/managed-resources/](https://docs.crossplane.io/latest/managed-resources/managed-resources/)  
   Explains MRs in general, not AWS-specific options.

5. **Creating YAML yourself (from Terraform docs or CRD)**  
   See [Creating Crossplane YAML from Docs](./Creating_Crossplane_YAML_From_Docs.md): `kubectl create ... --dry-run=client -o yaml`, `kubectl explain`, Terraform doc as reference (snake_case → camelCase), and optional `terraform providers schema -json` for custom automation.

---

## See also: reusable modules vs Configurations

For **how Terraform modules relate to Crossplane Configurations** (why generic mega-modules fit Terraform differently, curated XRDs, fork vs `tfvars`), see **[Terraform modules vs Crossplane Configurations](./20_Terraform_Modules_vs_Crossplane_Configurations.md)**.

---

## References

- [Terraform AWS Provider – Resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources)
- [Upbound Marketplace – Providers](https://marketplace.upbound.io/providers)
- [Upbound provider-aws-eks (Managed Resources)](https://marketplace.upbound.io/providers/upbound/provider-aws-eks/latest)
- [doc.crds.dev – crossplane/provider-aws](https://doc.crds.dev/github.com/crossplane/provider-aws)
- [Crossplane – Managed Resources](https://docs.crossplane.io/latest/managed-resources/managed-resources/)
- [Crossplane vs Terraform (Spacelift)](https://spacelift.io/blog/crossplane-vs-terraform)
- [Crossplane vs Terraform (Crossplane blog)](https://blog.crossplane.io/crossplane-vs-terraform)
