# Crossplane Course Curriculum – Study Guide

## Overview

This curriculum organizes study materials for **Crossplane v2**, the cloud-native control plane for platform engineering. Each section links to detailed guides with explanations, v2-specific content, and practical examples.

**Official docs:** [https://docs.crossplane.io/latest/](https://docs.crossplane.io/latest/)  
**GitHub:** [https://github.com/crossplane](https://github.com/crossplane)

---

## Course Structure

| # | Topic | Guide | Focus |
|---|--------|--------|--------|
| 1 | What is Crossplane? | [📘 What is Crossplane](./guides/01_What_Is_Crossplane.md) | Concepts, why control planes, architecture; **intuition:** framework for *your* APIs (Go analogy) |
| 2 | What's New in v2 | [🆕 What's New in v2](./guides/02_Whats_New_In_V2.md) | Namespaced XRs/MRs, compose any resource, Operations |
| 3 | Installation & Setup | [🛠️ Installation & Setup](./guides/03_Installation_Setup.md) | Helm install, feature flags, providers |
| 4 | Managed Resources & Providers | [☁️ Managed Resources](./guides/04_Managed_Resources_Providers.md) | Providers, MRs, ProviderConfig, RBAC |
| 5 | Composition (XRDs, Compositions, Functions) | [🧩 Composition](./guides/05_Composition_XRDs_Functions.md) | XRDs, Compositions, Patch & Transform, CUE/KCL/Python |
| — | **Crossplane for K8s applications (v2)** | [🚀 K8s Applications](./guides/09_Crossplane_For_Kubernetes_Applications.md) | Compose Deployments, Services, etc. (not only cloud MRs) |
| 6 | Operations | [⏱️ Operations](./guides/06_Operations.md) | CronOperation, WatchOperation, day-two workflows |
| 7 | Packages & CLI | [📦 Packages & CLI](./guides/07_Packages_CLI.md) | Functions, Configurations, crossplane CLI, render |
| — | **Upbound Marketplace** | [🏪 Upbound Marketplace](./guides/11_Upbound_Marketplace.md) | Why the Marketplace matters; Providers vs Configurations; discover & install packages; configuration-aws-eks |
| — | **AWS family: contrib vs Upbound** | [⚠️ contrib vs Upbound](./guides/18_AWS_Family_Contrib_vs_Upbound.md) | Why two family providers conflict; one stack per cluster; cleanup ProviderRevisions |
| 8 | Upgrade & Troubleshooting | [🔧 Upgrade & Troubleshooting](./guides/08_Upgrade_Troubleshooting.md) | v1→v2 upgrade, common issues |
| 9 | **GitOps** | [🔄 GitOps](./guides/10_GitOps.md) | Argo CD, Flux, directory structure, apply order, drift |
| — | **Crossplane vs Terraform** | [⚖️ Crossplane vs Terraform](./guides/Crossplane_vs_Terraform.md) | Comparison + where to find “resource docs” (Terraform registry equivalent) |
| — | **TF modules vs Crossplane Configurations** | [📐 Terraform modules vs Configurations](./guides/20_Terraform_Modules_vs_Crossplane_Configurations.md) | Reuse patterns: curated XRDs, fork vs variables, EKS/Helm examples |
| — | **Creating YAML from docs** | [📝 Creating YAML from Docs](./guides/Creating_Crossplane_YAML_From_Docs.md) | Create Crossplane YAML from Terraform docs / kubectl explain / CRD |
| — | **Composition Functions (available & custom)** | [🔌 Functions: Available & Custom](./guides/Composition_Functions_Available_And_Custom.md) | Where function code lives, other functions (CUE, auto-ready), write your own (Go/Python) |
| — | **Configuration from a Helm chart** | [📦 Helm → Crossplane](./guides/19_Configuration_From_Helm_Chart.md) | `provider-helm` `Release`, XR + Composition vs per-object MRs; tie-in to EKS kubeconfig |

---

## Quick Links to All Guides

- 📘 [01 – What is Crossplane?](./guides/01_What_Is_Crossplane.md) – includes **Intuition: a framework for your APIs** (Go analogy)
- 🆕 [02 – What's New in v2](./guides/02_Whats_New_In_V2.md)
- 🛠️ [03 – Installation & Setup](./guides/03_Installation_Setup.md)
- ☁️ [04 – Managed Resources & Providers](./guides/04_Managed_Resources_Providers.md)
- 🧩 [05 – Composition (XRDs, Compositions, Functions)](./guides/05_Composition_XRDs_Functions.md)
- 🚀 [Crossplane for Kubernetes Applications](./guides/09_Crossplane_For_Kubernetes_Applications.md) – compose normal K8s apps (v2)
- ⏱️ [06 – Operations](./guides/06_Operations.md)
- 📦 [07 – Packages & CLI](./guides/07_Packages_CLI.md)
- 🏪 [11 – Upbound Marketplace](./guides/11_Upbound_Marketplace.md) – why it matters; Providers vs Configurations; discover & install; configuration-aws-eks
- ⚠️ [AWS family: contrib vs Upbound](./guides/18_AWS_Family_Contrib_vs_Upbound.md) – do not mix two family providers; errors and cleanup
- 🔧 [08 – Upgrade & Troubleshooting](./guides/08_Upgrade_Troubleshooting.md)
- 🔄 [10 – GitOps](./guides/10_GitOps.md) – Argo CD, Flux, directory structure, apply order
- ⚖️ [Crossplane vs Terraform](./guides/Crossplane_vs_Terraform.md) – comparison & where to find Crossplane “resource docs” (Terraform registry equivalent)
- 📐 [20 – Terraform modules vs Crossplane Configurations](./guides/20_Terraform_Modules_vs_Crossplane_Configurations.md) – reuse: curated APIs vs generic modules, when fork beats `tfvars`
- 📝 [Creating YAML from Docs](./guides/Creating_Crossplane_YAML_From_Docs.md) – create Crossplane YAML from Terraform docs, kubectl explain, or CRD
- 🔌 [Composition Functions: Available & Custom](./guides/Composition_Functions_Available_And_Custom.md) – where function code lives, other functions, write your own (Go/Python)
- 📦 [19 – Configuration from a Helm chart](./guides/19_Configuration_From_Helm_Chart.md) – `provider-helm` Release, Crossplane vs Helm reconciliation

---

## Practical Examples

All runnable examples live in [`examples/`](./examples/):

| Example | Description | Try it |
|---------|-------------|--------|
| [00-official-doc-awsbucket](./examples/00-official-doc-awsbucket/) | Same layout as [Get Started With Managed Resources](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/) (contrib S3 + `ClusterProviderConfig`) | See folder README; apply provider → secret → config → bucket |
| [01-simple-s3-bucket](./examples/01-simple-s3-bucket/) | Create an S3 bucket with provider-aws-s3 | `kubectl apply -f bucket.yaml` |
| [02-provider-config-aws](./examples/02-provider-config-aws/) | ProviderConfig + ClusterProviderConfig, credentials | See README |
| [03-xrd-composition-bucket](./examples/03-xrd-composition-bucket/) | Custom API (XRD) + Composition → S3 bucket | `crossplane render` or apply |
| [04-patch-and-transform](./examples/04-patch-and-transform/) | Function Patch & Transform, FromCompositeFieldPath | `crossplane render` |
| [05-app-with-deployment](./examples/05-app-with-deployment/) | v2: Compose Deployment + Service (any K8s resource) | Apply XRD + Composition + XR |
| [06-cron-operation](./examples/06-cron-operation/) | CronOperation for scheduled tasks (alpha) | Enable Operations feature |
| [07-multi-resource-composition](./examples/07-multi-resource-composition/) | One XR → multiple composed resources (Bucket) | Full pipeline example |
| [08-application](./examples/08-application/) | **Application:** Kubernetes app (XApp → Deployment + Service) | Apply XRD + Composition + XR |
| [configuration-from-helm-chart](./examples/configuration-from-helm-chart/) | **Helm via Crossplane:** sample `Release` MR (`provider-helm`) | Install provider-helm + ProviderConfig; see [guide 19](./guides/19_Configuration_From_Helm_Chart.md) |
| [cue-xrd-composition](./examples/cue-xrd-composition/) | Cue → generated XRD/Composition in `out/` | `./build.sh` then apply or use with `crossplane render` |

---

## Configurations & real-world EKS

These folders use **Upbound Marketplace** packages and show the difference between **Configurations** (one XR) and **provider-based** (full control):

| Folder | Description | Try it |
|--------|-------------|--------|
| [configuration-aws-eks](./configuration-aws-eks/) | **Configuration:** One **EKS** XR from [Upbound configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks); composes cluster + node group + IAM + VPC | Install Configuration, then `kubectl apply -f eks-xr-example.yaml` |
| [configuration-openforms](./configuration-openforms/) | **Configuration:** **OpenForms** app — composed K8s resources (no Helm at runtime); port of [maykinmedia/openforms](https://github.com/maykinmedia/charts/tree/main/charts/openforms) chart | P&T + auto-ready; XR name must be `demo`; see folder README |
| [eks-cluster-spot](./eks-cluster-spot/) | **Provider-based:** Raw EKS, IAM, EC2 manifests; Spot node group, VPC/subnets; full control | Apply providers, IAM, VPC, subnets, cluster, nodegroup (see README) |

See [Upbound Marketplace](./guides/11_Upbound_Marketplace.md) for why the Marketplace matters and when to use a Configuration vs providers.

---

## Learning Path

1. **Basics:** 01 What is Crossplane → 03 Installation → 04 Managed Resources (try `00-official-doc-awsbucket` to mirror the official doc, or `01-simple-s3-bucket`).
2. **Abstractions:** 05 Composition → try `03-xrd-composition-bucket` and `04-patch-and-transform`.
3. **Marketplace & Configurations:** 11 Upbound Marketplace → understand Providers vs Configurations; try [configuration-aws-eks](./configuration-aws-eks/) (one EKS XR) and compare with [eks-cluster-spot](./eks-cluster-spot/) (provider-based).
4. **K8s applications (v2):** 09 Crossplane for Kubernetes Applications → try `05-app-with-deployment` (compose Deployment + Service); for a full app port without Helm at runtime, see [configuration-openforms](./configuration-openforms/).
5. **Day-two:** 06 Operations → try `06-cron-operation` (if enabled).
6. **GitOps:** 10 GitOps → Argo CD or Flux with Crossplane (directory structure, apply order).
7. **Production:** 07 Packages & CLI, 08 Upgrade & Troubleshooting.

---

## References

- [Upbound Marketplace](./guides/11_Upbound_Marketplace.md) – **why the Marketplace matters**; Providers vs Configurations; discover & install packages; [configuration-aws-eks](./configuration-aws-eks/) and [eks-cluster-spot](./eks-cluster-spot/)
- [AWS family: contrib vs Upbound](./guides/18_AWS_Family_Contrib_vs_Upbound.md) – **one family per cluster**; ProviderRevision conflicts; cleanup
- [GitOps with Crossplane](./guides/10_GitOps.md) – Argo CD, Flux, directory structure
- [Crossplane vs Terraform](./guides/Crossplane_vs_Terraform.md) – comparison; **where to find Crossplane AWS resource docs** (Terraform registry equivalent: Upbound Marketplace, doc.crds.dev, `kubectl explain`)
- [Crossplane Docs (latest)](https://docs.crossplane.io/latest/)
- [What's New in v2](https://docs.crossplane.io/latest/whats-new/)
- [Install Crossplane](https://docs.crossplane.io/latest/get-started/install/)
- [Composition – Composite Resources](https://docs.crossplane.io/latest/composition/composite-resources/)
- [Composition – Compositions](https://docs.crossplane.io/latest/composition/compositions/)
- [Packages – Providers](https://docs.crossplane.io/latest/packages/providers/)
- [Upgrade to Crossplane v2](https://docs.crossplane.io/latest/guides/upgrade-to-crossplane-v2/)
- [Crossplane GitHub](https://github.com/crossplane)
