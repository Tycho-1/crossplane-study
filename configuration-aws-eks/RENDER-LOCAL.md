# Generate composed YAML from a Configuration (no cluster apply)

You **cannot** get a full “Terraform plan” from Crossplane, but you **can** **render** what a **Composition pipeline** (functions) would emit as **YAML** using the **Crossplane CLI** — **without** `kubectl apply` to a live control plane.

**Important:** This still **runs composition functions locally in Docker** (they pull OCI images). It does **not** create AWS resources; it only prints **desired** Kubernetes manifests (managed resources + updated composite status).

**Docs:** [Crossplane CLI — `crossplane render`](https://docs.crossplane.io/latest/cli/command-reference/#render)

---

## Prerequisites

- **`crossplane` CLI** installed ([install](https://docs.crossplane.io/latest/cli/)).
- **Docker** running (`crossplane render` uses Docker to execute function containers).

---

## What you need on disk

| Input | Purpose |
|--------|---------|
| **XR YAML** | Your composite instance, e.g. [eks-xr-example.yaml](./eks-xr-example.yaml) (`kind: EKS`, `spec.parameters`, `metadata.name` / `namespace`). |
| **Composition YAML** | The pipeline for that composite — from the package source, e.g. upstream [`apis/eks/composition.yaml`](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/composition.yaml). |
| **`functions.yaml`** | One or more `kind: Function` documents so the CLI knows **which OCI packages** to run — **same `spec.package` references** as on a real cluster (see below). |

There is **no** need to have Crossplane installed on a cluster for `crossplane render` itself.

---

## Basic command

```bash
crossplane render eks-xr-example.yaml composition.yaml functions.yaml
```

Pipe or redirect to a file:

```bash
crossplane render eks-xr-example.yaml composition.yaml functions.yaml > rendered.yaml
```

Useful flags (see `crossplane render --help`):

| Flag | Use |
|------|-----|
| `-x` / `--include-full-xr` | Include full XR **spec** + **metadata** in output (not only status). |
| `-r` / `--include-function-results` | Include function **results** / events in output. |
| `--observed-resources=DIR` | **Mock** already-“existing” composed resources so functions that branch on readiness see them (advanced; see caveat below). |

---

## `functions.yaml` for **configuration-aws-eks**

The EKS Composition pipeline has **two** steps ([upstream `composition.yaml`](https://github.com/upbound/configuration-aws-eks/blob/main/apis/eks/composition.yaml)):

1. **`upbound-configuration-aws-ekseks`** → package **`configuration-aws-eks_eks`**
2. **`crossplane-contrib-function-auto-ready`**

Your `functions.yaml` should declare **`Function`** objects whose **`metadata.name`** matches **`functionRef.name`** in the Composition, and whose **`spec.package`** matches what you install in the cluster.

**Copy exact `spec.package` from a healthy cluster:**

```bash
kubectl get function upbound-configuration-aws-ekseks -o yaml
kubectl get function crossplane-contrib-function-auto-ready -o yaml
```

Minimal **template** (replace `package:` lines with your cluster’s values / a pinned tag):

```yaml
---
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: upbound-configuration-aws-ekseks
spec:
  package: xpkg.upbound.io/upbound/configuration-aws-eks_eks:<tag-or-digest>
---
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: crossplane-contrib-function-auto-ready
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-auto-ready:<tag>
```

Then run:

```bash
crossplane render eks-xr-example.yaml apis/eks/composition.yaml functions.yaml
```

(paths adjusted to where you cloned [configuration-aws-eks](https://github.com/upbound/configuration-aws-eks))

---

## Caveat: EKS KCL uses **readiness gates**

Upstream [`functions/eks/main.k`](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k) only adds some resources **after** others appear “ready” in **observed** state (e.g. `ClusterAuth` after cluster, `NodeGroup` after `vpc-cni-addon` exists, etc.).

On a **first** `crossplane render` with an **empty** observed set, you may see **only a subset** of the final composed resources — similar to “first reconcile” vs “steady state.”

To approximate later passes locally, use **`--observed-resources`** with YAML files that mock resources the function expects (cluster status, addon ready, …). That is **manual** and brittle; most people use **render** for simpler compositions or for debugging one pipeline step.

---

## What does **not** replace `render`

| Approach | What it does |
|----------|----------------|
| **`kubectl apply --dry-run=client -o yaml`** | Validates / prints your **XR** manifest only — **not** composed children. |
| **`kubectl apply --dry-run=server`** | Needs a cluster API; still **not** the function output. |
| **Reading `main.k`** | Source of truth for logic; does not by itself emit YAML without running KCL/function. |

---

## Upbound **`up` CLI**

From the configuration repo, **`up project run`** / **`up project build`** are aimed at **building and running** the project (often against a dev control plane). For **offline YAML preview** of composed resources, prefer **`crossplane render`** unless your Upbound workflow documents a specific render target.

See also [guides/17_Up_CLI_And_DevEx.md](../guides/17_Up_CLI_And_DevEx.md).

---

## See also

- [CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md) — XR → Composition → Function → image → `main.k`
- [HOW-IT-WORKS.md](./HOW-IT-WORKS.md) — what the EKS function creates at runtime
- [examples/03-xrd-composition-bucket/README.md](../examples/03-xrd-composition-bucket/README.md) — smaller example mentioning `crossplane render`
