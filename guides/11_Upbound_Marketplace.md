# Upbound Marketplace – Why It Matters

The **Upbound Marketplace** is the main place to discover and install Crossplane **packages**: Providers, Configurations, and Functions. Understanding it is important for your Crossplane course because almost every “how do I create X on AWS?” question leads there.

**Marketplace:** [https://marketplace.upbound.io](https://marketplace.upbound.io)

---

## What is the Upbound Marketplace?

The Marketplace is **Upbound’s package registry** for Crossplane. It is roughly the Crossplane equivalent of:

- Terraform’s **Registry** (registry.terraform.io) for modules and providers  
- Docker Hub for images  
- Helm’s artifact hub for charts  

You use it to:

1. **Discover** what packages exist (AWS EKS, S3, EC2, Configurations that compose whole clusters, etc.).
2. **Choose versions** (e.g. `v1.2.0`, `latest`) and see compatibility (Crossplane / UXP version).
3. **Get install manifests** – copy-paste YAML to install a Provider or Configuration in your cluster.
4. **Read API docs** – Managed Resources, parameters, and (for Configurations) the custom API (XRD) and what gets composed.

So: **when you need “the Crossplane way to do EKS / S3 / RDS,” the Marketplace is where you start.**

---

## Why the Marketplace matters for the course

- **One place for “what exists”** – Instead of hunting GitHub or docs, you go to [marketplace.upbound.io](https://marketplace.upbound.io) → Providers or Configurations, and browse or search.
- **Correct package names and versions** – Install manifests use full package URLs (e.g. `xpkg.upbound.io/upbound/provider-aws-eks:v1.13.1`). The Marketplace shows these so you don’t guess.
- **Configurations vs Providers** – The Marketplace separates **Providers** (low-level CRDs: Cluster, Bucket, etc.) from **Configurations** (higher-level APIs: one “EKS” XR that composes cluster + node group + IAM + VPC). That distinction is central to platform design.
- **Compatibility** – Package pages often state compatibility (e.g. Crossplane 2.0+, UXP 2.0+). You need that when setting up your lab or recommending versions.

Without the Marketplace, you’d be digging through GitHub repos and CRD YAML to find package names, versions, and schemas. **With it, the course can point students to a single, stable entry point.**

---

## What’s on the Marketplace?

| Section | What it is | Example |
|--------|------------|--------|
| **Providers** | Packages that add **Managed Resources** (CRDs + controller) for a cloud or system | [provider-aws-eks](https://marketplace.upbound.io/providers/upbound/provider-aws-eks), [provider-aws-s3](https://marketplace.upbound.io/providers/upbound/provider-aws-s3), [provider-aws-ec2](https://marketplace.upbound.io/providers/upbound/provider-aws-ec2) |
| **Configurations** | Packages that add **XRDs + Compositions** (and often embedded functions); you get a **custom API** (e.g. one “EKS” resource) that composes many underlying resources | [configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks) |
| **Functions** | Composition **functions** used in pipelines (Patch & Transform, CUE, auto-ready, etc.) | [function-patch-and-transform](https://marketplace.upbound.io/functions/crossplane-contrib/function-patch-and-transform), [function-auto-ready](https://marketplace.upbound.io/functions/crossplane-contrib/function-auto-ready) |

- **Providers** = “I want to create raw AWS/API resources” (Cluster, NodeGroup, Bucket, VPC, Subnet, …).  
- **Configurations** = “I want a higher-level API; one resource (e.g. EKS) that creates cluster + node group + IAM + VPC.”  
- **Functions** = “I need a function for my Composition pipeline.”

---

## Providers: finding and installing

1. Go to [marketplace.upbound.io/providers](https://marketplace.upbound.io/providers).
2. Search or filter (e.g. “aws”, “eks”).
3. Open a provider (e.g. **provider-aws-eks**) → choose a **version** (e.g. v1.13.1).
4. Use the **Install Manifest** (or copy the `Provider` YAML). Example:

   ```yaml
   apiVersion: pkg.crossplane.io/v1
   kind: Provider
   metadata:
     name: provider-aws-eks
   spec:
     package: xpkg.upbound.io/upbound/provider-aws-eks:v1.13.1
   ```

5. Apply; wait until the provider is HEALTHY. Then use the **Managed Resources** listed on the same page (e.g. `Cluster`, `NodeGroup`) and optional **API reference** links.

See also [Crossplane vs Terraform](./Crossplane_vs_Terraform.md) for “where to find Crossplane AWS resource docs” (Marketplace → provider → Managed Resources; doc.crds.dev; `kubectl explain`).

---

## Configurations: higher-level APIs

A **Configuration** is a package that ships **XRDs + Compositions** (and often **embedded functions**). It does **not** only add low-level CRDs; it adds a **custom API** (Composite Resource type) and the logic to turn one high-level resource into many low-level ones.

### Why Configurations matter

- **Simpler UX** – Users create one resource (e.g. `EKS` with id, region, node count) instead of many (Cluster, NodeGroup, IAM roles, VPC, subnets).
- **Curated patterns** – The Configuration author decides the schema and defaults (version, instance types, networking). Good for platform teams offering a “standard” EKS.
- **Discovery** – On the Marketplace you can browse [Configurations](https://marketplace.upbound.io/configurations) and see “EKS Configuration,” “S3 Configuration,” etc., and choose between “raw provider” vs “ready-made composition.”

### Example: configuration-aws-eks

- **Marketplace:** [upbound/configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks/v2.0.0)  
- **What it does:** Defines an **EKS** custom resource (`aws.platform.upbound.io/v1alpha1`). When you create one EKS XR, the Configuration’s Composition (and embedded function) provisions EKS Cluster, NodeGroup, IAM, and typically VPC/subnets.  
- **In this repo:** See the **[configuration-aws-eks](../configuration-aws-eks/)** folder: install manifest (`configuration-aws-eks-install.yaml`) and example XR (`eks-xr-example.yaml`). That’s the “configuration way” to create a cluster.

Contrast with the **[eks-cluster-spot](../eks-cluster-spot/)** folder: raw provider manifests (Cluster, NodeGroup, IAM, VPC, subnets) and full control (e.g. Spot capacity). That’s the “provider way.”

---

## How to install a Configuration

1. Open the Configuration on the Marketplace (e.g. [configuration-aws-eks v2.0.0](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks/v2.0.0)).
2. Use the **Install Manifest** (or equivalent). It looks like:

   ```yaml
   apiVersion: pkg.crossplane.io/v1
   kind: Configuration
   metadata:
     name: configuration-aws-eks
   spec:
     package: xpkg.upbound.io/upbound/configuration-aws-eks:v2.0.0
   ```

3. Apply; wait until the **Configuration** is HEALTHY (`kubectl get configuration`).
4. Create an instance of the **custom API** (e.g. an `EKS` XR). The Composition will create the underlying resources.

Package dependencies (providers, functions) are often pulled in by the Configuration; check the Marketplace page for “Dependencies” or install notes.

---

## Where this fits in the curriculum

- **After “Packages & CLI”** – You already know that Crossplane has Providers, Functions, and Configurations. The Marketplace is **where you get them**: names, versions, install YAML.
- **Before or with “Composition”** – Configurations are “Composition as a product”: someone else wrote the XRD and Composition; you just install the package and create XRs.
- **With “Managed Resources & Providers”** – When you use a provider (e.g. EKS), the Marketplace is where you look up which package to install and which Managed Resources exist.
- **Real examples in this repo:**
  - **[configuration-aws-eks](../configuration-aws-eks/)** – One XR (EKS) from the Marketplace Configuration.
  - **[eks-cluster-spot](../eks-cluster-spot/)** – Raw provider manifests (EKS, IAM, EC2) for full control (e.g. Spot).

---

## Summary

| Idea | Takeaway |
|------|----------|
| **What** | Upbound Marketplace = package registry for Providers, Configurations, and Functions. |
| **Why** | Single place to discover packages, versions, install manifests, and API/schema info. |
| **Providers** | Low-level CRDs (Cluster, Bucket, VPC, …); install via `Provider`; use when you want full control. |
| **Configurations** | High-level API (e.g. one “EKS” XR) + Compositions; install via `Configuration`; use when you want a simple, curated API. |
| **Course** | Point students to the Marketplace early; use it for “how do I do EKS/S3/…” and to contrast provider vs configuration (e.g. eks-cluster-spot vs configuration-aws-eks). |

**Links**

- [Upbound Marketplace](https://marketplace.upbound.io)
- [Providers](https://marketplace.upbound.io/providers)
- [Configurations](https://marketplace.upbound.io/configurations)
- [configuration-aws-eks (this repo)](../configuration-aws-eks/)
- [eks-cluster-spot – provider-based EKS (this repo)](../eks-cluster-spot/)
