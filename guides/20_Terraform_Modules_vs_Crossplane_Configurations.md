# Terraform modules vs Crossplane Configurations

This guide explains why **reusable Terraform modules** (e.g. a generic EKS or VPC module) **do not map one-to-one** to **Crossplane Configurations** (packages with an XRD + Composition). It complements the high-level tool comparison in **[Crossplane vs Terraform](./Crossplane_vs_Terraform.md)** and the **curated API** idea in **[configuration-aws-eks/XR-PARAMETERS.md](../configuration-aws-eks/XR-PARAMETERS.md#curated-eks-api-vs-full-terraform-surface)** and **[configuration-openforms/USAGE.md](../configuration-openforms/USAGE.md#design-curated-parameters-vs-full-helm-values)**.

**Related intuition:** Crossplane is a **framework for your own APIs** (like Go is a language for your own programs) — see **[What is Crossplane? — Intuition](./01_What_Is_Crossplane.md#intuition-a-framework-for-your-apis)**.

---

## Terraform module vs Crossplane “module”

| | **Terraform** | **Crossplane** |
|---|----------------|----------------|
| **Typical reusable unit** | A **module** (HCL): many `variable` blocks; callers supply **`.tfvars`**. | A **Configuration** (OCI package): **CompositeResourceDefinition (XRD) + Composition** — a **custom API** plus **how** it is implemented. |
| **“Generic EKS module”** | Common pattern: **one** module, **many** teams, **different variables** per environment. | A **vendor Configuration** (e.g. [configuration-aws-eks](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks)) behaves more like a **product**: **fewer** top-level fields and **strong opinions** baked into the Composition / functions. |
| **What you customize** | Mostly **values** (variables, `tfvars`, workspace locals). | **Schema** (what exists on the XR) **and** **behavior** (Composition, functions). Extending behavior often means **changing** the XRD or **forking** the Configuration — not only changing defaults. |

Configurations are **not** “HCL with different syntax”: the **composite API** (`spec.parameters`) is the **contract** your platform exposes.

---

## Why a “big generic module” fits Terraform more naturally

- In Terraform, **flexibility usually lives in variables**: the **same** module can serve many companies with different `tfvars`, without forking the module.
- In Crossplane, **flexibility in a packaged Configuration** is largely **whatever the XRD exposes**. If the XR only has `id`, `region`, `nodes`, … then **every consumer** of that package shares that **same** API. Adding knobs means **extending the XRD**, **forking** the Configuration, or maintaining a **different** Composition — not merely “more tfvars.”

So the pattern **“one important shared module (EKS, VPC, …), hundreds of variables, every org passes different values”** maps **less directly** to Crossplane than to Terraform: published Configurations tend to be **more opinionated** and **less** like “parameterize every cloud attribute.”

---

## What *does* make sense for Crossplane

| Layer | Role |
|-------|------|
| **Provider packages** | Broad, shared (e.g. AWS family) — many **managed resources** (MRs), same idea as Terraform **providers**. |
| **Vendor / Marketplace Configurations** | **Starting points**: opinionated **EKS**, **Network**, etc. Many organizations **fork** and adjust XRD/Composition to match **their** platform (naming, networking, IAM, allowed regions). |
| **In-house Configurations** | Your **own** small API (`spec.parameters`) on purpose; that API **is** your internal platform product — not a thin pass-through of every AWS or Helm field. |

Reuse in Crossplane often looks like **“install provider + fork or wrap a Configuration”** more than **“one generic module, infinite variables.”**

---

## Every organization chooses what fits

Same as with Terraform (which modules to use, fork, or wrap), but the **artifact** differs:

| | **Terraform** | **Crossplane** |
|---|----------------|----------------|
| **Typical customization** | Reuse a registry module; vary **variables** and **workspace** state. | Reuse a Configuration **or** **fork** it so **your** XR schema matches **your** governance and golden paths. |
| **“Our EKS” as a product** | Often a **wrapper module** around a public EKS module with **company defaults** in HCL. | Often a **custom XRD** + **Composition** (or a **forked** vendor Configuration) so **claims** only allow approved shapes. |

There is **no** single “right” degree of opinion: **curated** APIs vs **full** surface area is a **product** decision (see also [USAGE.md — Design: curated parameters vs full Helm values](../configuration-openforms/USAGE.md#design-curated-parameters-vs-full-helm-values)).

---

## Summary

| Question | Short answer |
|----------|----------------|
| Are Crossplane Configurations the same as Terraform modules? | **No.** A Configuration is an **API + implementation** (XRD + Composition), not only parameterized HCL. |
| Is a mega generic “EKS module” as central for Crossplane as for Terraform? | **Usually less.** Terraform modules often maximize **variables**; Crossplane packages often **intentionally** expose a **smaller** XR surface. |
| Is reuse irrelevant for Crossplane? | **No** — **Providers** and **Configurations** are still shared — but reuse is often **fork + adapt** or **thin platform API**, not **one module, infinite tfvars**. |
| Who decides how opinionated to be? | **Each organization** — same as choosing Terraform module granularity, but the **levers** are XRD shape and Composition logic, not only `tfvars`. |
| Do Crossplane + Helm + Terraform all stay relevant? | **Usually yes** — **different usage**, often **coexistence**; Crossplane **partially** overlaps some workflows only where you **choose** it (see next section). |
| Is there “Upjet for Terraform modules”? | **No** in the same sense as **[Upjet](https://github.com/crossplane/upjet) for providers** — modules are not provider schemas; see [Upjet vs modules](#upjet-terraform-providers-vs-terraform-modules). |
| Should a composite API mirror every module variable? | **Usually no** — **curate** the XR; blind import of all variables defeats a platform API. |

---

## Helm, Terraform, and Crossplane: complement, not wholesale replacement

Crossplane **does not** make **Helm** or **Terraform** obsolete in general. Many real deployments use **all three** in different roles:

| Tool | Typical role |
|------|----------------|
| **Terraform** | Bootstrap accounts, networking, IAM; multi-cloud or **non-Kubernetes** IaC; teams standardized on `terraform apply` and remote **state**. |
| **Helm** | Kubernetes **packaging**, chart ecosystem, **upgrade/rollback** semantics; vendors often ship **only** charts. |
| **Crossplane** | **Desired state** as Kubernetes **CRs**, continuous **reconciliation**, **platform APIs** (XRDs + Compositions), GitOps on those APIs, **managed resources** for cloud APIs. |

**Different usage, not a drop-in substitute:** Terraform is **CLI + state**; Helm is a **release manager** for Kubernetes; Crossplane is a **control plane running in the cluster**. They solve overlapping problems in **different** ways.

**Nuance — partial overlap:** Crossplane **can** replace **some** workflows you might otherwise do with Helm or Terraform — for example a **[`provider-helm` `Release`](https://github.com/crossplane-contrib/provider-helm)** instead of a separate Helm CI/CD path for certain apps, or **AWS MRs** instead of Terraform for selected resources — but that is an **explicit architecture choice**, not proof that “Crossplane replaces Helm and Terraform” everywhere.

**Practical conclusion:** **Helm and Terraform stay relevant** for many teams and pipelines. Crossplane adds a **Kubernetes-native control-plane** layer. **Coexistence** is the common pattern; **full replacement** of both tools is the exception. See also [Configuration from a Helm chart](./19_Configuration_From_Helm_Chart.md) for when Helm **`Release`** still makes sense inside Crossplane.

---

## Upjet (Terraform providers) vs Terraform modules

**[Upjet](https://github.com/crossplane/upjet)** generates Crossplane **provider** code from a **Terraform provider** schema: Terraform **resource types** (e.g. `aws_eks_cluster`) become **managed resources** (CRDs + controllers). That pipeline is **machine-oriented** — provider schemas are **stable and structured**.

| | **Terraform provider** | **Terraform module** |
|---|------------------------|----------------------|
| **What it defines** | **Resource types** and their arguments (provider schema). | **Arbitrary HCL**: multiple resources, `variable`, `locals`, `count` / `for_each`, outputs — **different patterns per module**. |
| **Codegen fit** | **[Upjet](https://github.com/crossplane/upjet)** maps **resource type → MR** in a repeatable way. | **No** standard “Upjet for modules” that turns any registry module into a finished **Configuration** the same way. |
| **Crossplane analogue** | **Provider package** (many MRs). | **XRD + Composition** (your **platform API**) — usually **authored** or **forked**, not generated from arbitrary module IR. |

So **by definition** the problem is different: **providers** are one abstraction layer; **modules** are **composed stacks + opinions**.

### Why there is no generic “Upjet for modules” (and why you would not want a blind import)

1. **No single schema for “all modules”** — unlike a provider schema, each module is a **custom** graph of resources and logic.
2. **A good XR is a *subset*** — you rarely want every `variable` from a fat module on the composite API; **curation** (what [curated EKS](../configuration-aws-eks/XR-PARAMETERS.md#curated-eks-api-vs-full-terraform-surface) and [Open Forms USAGE](../configuration-openforms/USAGE.md#design-curated-parameters-vs-full-helm-values) describe) is **product design**, not something to auto-dump from HCL.
3. **Partial tooling** (templates, generators, AI-assisted YAML) can still help, but **someone** must decide **`spec.parameters`** — automation does not remove that step.

**Bottom line:** **Upjet-style codegen fits Terraform *providers***. **Terraform *modules*** map to **Composition + XRD**, where **selection** is intentional — so there is **no** equivalent universal “Upjet for modules,” and **importing every variable** would usually be the wrong outcome anyway.

---

## See also

- **What is Crossplane? (Go-style intuition):** [01 – What is Crossplane?](./01_What_Is_Crossplane.md#intuition-a-framework-for-your-apis)
- **Crossplane vs Terraform (tools, docs):** [Crossplane](./Crossplane_vs_Terraform.md) — includes a short **“Do Crossplane and Terraform replace each other?”** note linking here
- **Configurations concept:** [Configurations](./13_Configurations.md)
- **EKS XR (curated vs Terraform):** [configuration-aws-eks/XR-PARAMETERS.md](../configuration-aws-eks/XR-PARAMETERS.md#curated-eks-api-vs-full-terraform-surface)
- **Helm chart vs curated XR (Open Forms):** [configuration-openforms/USAGE.md](../configuration-openforms/USAGE.md#design-curated-parameters-vs-full-helm-values)
- **Upbound Marketplace:** [Upbound Marketplace](./11_Upbound_Marketplace.md)
- **Upjet (providers from Terraform schemas):** [github.com/crossplane/upjet](https://github.com/crossplane/upjet) — not a tool for Terraform **modules**; see [Upjet vs modules](#upjet-terraform-providers-vs-terraform-modules) above
