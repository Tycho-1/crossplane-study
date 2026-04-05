# What is Crossplane?

## Introduction

**Crossplane** is a **control plane framework for platform engineering**. It extends Kubernetes so you can provision and manage **cloud infrastructure** and, in **v2**, **normal Kubernetes applications** (Deployments, Services, ConfigMaps, etc.) using the same APIs and workflows.

- **Kubernetes-native:** CRDs, controllers, `kubectl`, GitOps.
- **Cloud-agnostic:** AWS, Azure, GCP, and many providers via the package ecosystem.
- **Declarative:** Infrastructure and applications as YAML; reconcile to desired state.
- **Self-service:** Developers use custom APIs (Composite Resources) defined by the platform team.

**Two main use cases:**

1. **Cloud infrastructure** – Provision and manage cloud resources (S3, RDS, EKS, etc.) via Crossplane **managed resources** (MRs) and **Providers**.
2. **Kubernetes applications** (v2) – Create and manage normal K8s workloads (Deployments, Services, ConfigMaps, etc.) via **Compositions** that compose any Kubernetes resource, not only MRs. See [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md).

### Intuition: a framework for your APIs

A loose analogy: **Go** is a language — you can build many kinds of programs; **you** choose packages and structure. **Crossplane** is a **framework** — you define **your own** platform APIs (**XRDs**), how they are **implemented** (**Compositions**, functions), and what gets composed. There is no single fixed “Crossplane app”; each team or vendor package exposes **different** composite kinds (`EKS`, `OpenForms`, `MyDatabase`, …).

You still work inside **Kubernetes** (CRDs, reconciliation, RBAC), but the **surface API** you offer to developers is **yours** to design — similar in spirit to “here is Go, build what you need” / “here is Crossplane, shape the control plane you want.” For how that differs from reusing giant Terraform **modules** or Helm **values** wholesale, see [Terraform modules vs Crossplane Configurations](./20_Terraform_Modules_vs_Crossplane_Configurations.md).

---

## Why Control Planes?

A **control plane** is the layer that decides *what* should exist and *how* it should look. Kubernetes is a control plane for containers; Crossplane turns your cluster into a control plane for **infrastructure and applications**.

- **Single API:** Request databases, buckets, networks, and (in v2) full apps via Kubernetes.
- **Abstraction:** Platform team defines XRs (e.g. `XDatabase`, `XApp`); users don’t touch cloud-specific CRDs.
- **Policy & governance:** RBAC, quotas, and composition logic enforce standards.

---

## Core Concepts

### 1. Managed Resources (MRs)

- **CRDs** that represent **external resources** (e.g. S3 Bucket, RDS instance).
- Provided by **Providers** (e.g. provider-aws, provider-azure).
- Crossplane (and the provider) **reconcile** MRs: create/update/delete the real cloud resource so that the MR’s spec matches reality.

**Example (v2 style – namespaced):**

```yaml
apiVersion: s3.aws.m.upbound.io/v1beta1
kind: Bucket
metadata:
  namespace: default
  name: my-bucket
spec:
  forProvider:
    region: eu-central-1
```

### 2. Providers

- **Packages** that install CRDs and a **controller** in the cluster.
- Controller watches MRs and talks to the cloud API (AWS, Azure, GCP, etc.).
- You configure **credentials** via ProviderConfig / ClusterProviderConfig (e.g. from a Secret).

### 3. Composition (XR + Composition + XRD)

- **XRD (Composite Resource Definition):** defines a **custom API** (group, kind, schema). Example: `mydatabases.example.org`, kind `MyDatabase`.
- **Composition:** defines *how* that API is implemented: which resources to create (MRs, or in v2 any K8s resources) and how to template them (e.g. via composition functions).
- **XR (Composite Resource):** an *instance* of that API. When you create an XR, Crossplane runs the matching Composition and creates the composed resources.

So: **XRD = API contract**, **Composition = implementation**, **XR = request**.

### 4. Composition Functions

- Compositions use a **pipeline** of **functions** (e.g. Patch & Transform, CUE, KCL, Python) to compute the set of composed resources from the XR.
- Replaces the legacy “patch and transform” embedded in the Composition spec; more flexible and testable (e.g. `crossplane render`).

---

## High-Level Architecture

```
User: kubectl apply -f xdatabase.yaml
         │
         ▼
Kubernetes API Server (XR / MR CRDs)
         │
         ▼
Crossplane Core
  - Picks Composition for XR
  - Calls composition functions → desired composed resources
  - Applies composed resources (MRs, Deployments, etc.)
         │
         ▼
Providers (e.g. AWS)
  - Watch MRs (Bucket, DBInstance, …)
  - Call cloud APIs (S3, RDS, …)
  - Update MR status
```

---

## What You’ll Do in This Course

1. **Install Crossplane** and a provider (e.g. provider-aws-s3).
2. **Create managed resources** (e.g. S3 bucket) and **ProviderConfig** for credentials.
3. **Define custom APIs** with XRDs and **implement** them with Compositions and functions (e.g. Patch & Transform).
4. Use **v2 features:** namespaced XRs/MRs, composing **any** Kubernetes resource (Deployments, Services, etc.).
5. Optionally use **Operations** (CronOperation, WatchOperation) for day-two workflows.

---

## Next

- [What's New in v2](./02_Whats_New_In_V2.md) – Namespaced resources, compose any resource, Operations.
- [Crossplane for Kubernetes Applications](./09_Crossplane_For_Kubernetes_Applications.md) – Using Crossplane for normal K8s apps (v2).
- [Installation & Setup](./03_Installation_Setup.md) – Install Crossplane and providers.
