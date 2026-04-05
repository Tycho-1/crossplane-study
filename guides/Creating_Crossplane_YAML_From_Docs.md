# Creating Crossplane YAML from Docs (Including Terraform Registry)

You want to create Crossplane managed resource YAML yourself. There is **no official “paste Terraform doc URL → Crossplane YAML”** tool, but you can do it in a repeatable way using Crossplane’s CRD schema and (optionally) Terraform docs as a reference.

---

## Can you automate YAML from the Terraform registry?

- **No end-user tool** that takes e.g. [Terraform’s eks_cluster doc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster) and outputs Crossplane YAML.
- **Upjet** generates Crossplane providers **from** Terraform provider schemas, so the **same options** exist in both; field names are usually **snake_case (Terraform) → camelCase (Crossplane)**. Upjet does **not** generate Crossplane APIs from Terraform **modules** (composed HCL) — that is a different layer; see **[Terraform modules vs Crossplane Configurations](./20_Terraform_Modules_vs_Crossplane_Configurations.md#upjet-terraform-providers-vs-terraform-modules)**.
- You can use Terraform docs to know **what** options exist, then get **exact field names** from Crossplane (see below).
- For **custom automation**, you can use `terraform providers schema -json` and build your own generator (see “Optional: Terraform provider schema” below).

---

## How to create YAML yourself from docs

### 1. From Crossplane (best once the provider is installed)

**Get a valid YAML skeleton:**

```bash
# List CRDs for the service (e.g. EKS)
kubectl get crd | grep eks

# Generate a YAML skeleton for a managed resource (no API call, no apply)
kubectl create cluster.eks.aws.upbound.io/v1beta1 Cluster my-eks \
  --dry-run=client -o yaml
```

That prints minimal valid YAML with `apiVersion`, `kind`, `metadata`, and empty `spec`.

**See all fields (types, descriptions):**

```bash
# All fields for the resource
kubectl explain cluster.eks.aws.upbound.io --recursive

# Or drill into spec
kubectl explain cluster.eks.aws.upbound.io.spec.forProvider
```

Use that output to fill `spec.forProvider` (and optional `providerConfigRef`, etc.).

---

### 2. Using Terraform docs as reference

- Open the Terraform resource page, e.g. [eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster).
- In the **Argument Reference**, each argument (e.g. `name`, `role_arn`, `vpc_config`, `version`) has a counterpart in Crossplane under `spec.forProvider`.
- Convert names to **camelCase**: `role_arn` → `roleArn`, `vpc_config` → `vpcConfig`, `subnet_ids` → `subnetIds`.
- Nested blocks in Terraform (e.g. `vpc_config { subnet_ids = [...] }`) become nested objects in YAML.

**Example mapping for EKS:**

| Terraform (eks_cluster)      | Crossplane (Cluster MR)              |
|-----------------------------|--------------------------------------|
| `name`                      | `spec.forProvider.name`              |
| `role_arn`                  | `spec.forProvider.roleArn`           |
| `version`                   | `spec.forProvider.version`           |
| `vpc_config { subnet_ids }` | `spec.forProvider.vpcConfig.subnetIds` |

So: Terraform doc tells you **what** options exist; **exact field names** come from `kubectl explain` or [doc.crds.dev](https://doc.crds.dev/github.com/crossplane/provider-aws).

---

### 3. From CRD YAML (no cluster)

If you have the CRD (e.g. from [provider-upjet-aws](https://github.com/crossplane-contrib/provider-upjet-aws) or [doc.crds.dev](https://doc.crds.dev/github.com/crossplane/provider-aws)), you can write a script that:

- Reads the CRD’s `openAPIV3Schema`
- Emits a YAML template with `spec.forProvider` keys and placeholder values

That’s custom tooling; there’s no standard CLI for it.

---

### 4. Optional: Terraform provider schema (for custom automation)

Terraform can dump the **provider schema as JSON** (resource names, argument types, nesting). You can use that to build your own “Terraform resource → Crossplane YAML” generator.

```bash
# In a directory with a minimal Terraform config that uses the AWS provider:
terraform init
terraform providers schema -json
```

See [terraform providers schema](https://developer.hashicorp.com/terraform/cli/commands/providers/schema).

You’d still need:

- A mapping from Terraform resource type (e.g. `aws_eks_cluster`) to Crossplane `apiVersion`/`kind` (e.g. `eks.aws.upbound.io/v1beta1`, `Cluster`)
- snake_case → camelCase for field names

---

## Summary: creating YAML yourself

| Method | When to use |
|--------|-------------|
| **`kubectl create ... --dry-run=client -o yaml`** | Provider already installed; you want a valid skeleton fast. |
| **`kubectl explain ... --recursive`** | To see exact field names and types for `spec.forProvider`. |
| **Terraform resource doc** | To know *what* options exist (same as Crossplane); convert argument names to camelCase for YAML. |
| **doc.crds.dev** | To see full CRD schema without a cluster. |
| **`terraform providers schema -json`** | If you want to build a custom “Terraform resource → Crossplane YAML” generator. |

---

## References

- [Terraform AWS Provider – eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster)
- [doc.crds.dev – crossplane/provider-aws](https://doc.crds.dev/github.com/crossplane/provider-aws)
- [Upjet](https://github.com/crossplane/upjet) (Crossplane providers from Terraform schemas)
- [terraform providers schema](https://developer.hashicorp.com/terraform/cli/commands/providers/schema)
- [Crossplane vs Terraform](./Crossplane_vs_Terraform.md) (comparison and where to find resource docs)
