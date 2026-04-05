# Crossplane RunFunction request / response model

Composition functions (KCL, CUE, Go, Python, Patch & Transform, …) all implement the same **gRPC contract**: Crossplane sends a **RunFunctionRequest**; the function returns a **RunFunctionResponse**. This doc summarizes that contract in plain language.

**Formal definitions** live in Crossplane’s protobufs (search for `RunFunctionRequest` / `RunFunctionResponse` in the [crossplane/crossplane](https://github.com/crossplane/crossplane) repo) and in the **[documentation for composition functions](https://docs.crossplane.io/latest/composition/compositions/)**.

---

## What triggers a function

When a **Composite Resource** (XR) is reconciled, Crossplane runs the **Composition pipeline** (ordered steps). Each step invokes one **Function** package. Each invocation is **one** request/response pair.

The function **does not call AWS/Azure/GCP**. It only returns **desired Kubernetes objects** (managed resources and optional XR patches). **Providers** reconcile those MRs to the cloud.

---

## Request — what Crossplane sends

Conceptual fields (exact names follow the protobuf / SDK version):

| Area | Meaning |
|------|--------|
| **Observed composite resource** | The **XR** as stored in the API: `metadata`, `spec` (e.g. `spec.parameters`), and current **status**. This is what KCL exposes as **`oxr`** (observed composite). |
| **Observed composed resources** | For each composed object the Composition manages: **current** manifest + **status** (e.g. provider conditions, `atProvider`). Lets logic branch on “is cluster ready?” — KCL **`_ocds`**. |
| **Desired composite / composed** | What **previous steps** (or the same step last time) asked for — your function can **update** or **replace** parts of desired state. |
| **Context** | Optional JSON blob (config, flags, or data from the platform / earlier steps). |
| **Resource requirements** (optional) | Some functions request **extra** cluster objects (e.g. via **extra-resources** patterns). |

Intuition: **“Here is the observed world and prior desired state — what should we desire next?”**

---

## Response — what the function returns

| Area | Meaning |
|------|--------|
| **Desired composed resources** | The **managed resources** Crossplane should apply (create/update/delete). **KCL** builds this list as **`items`**. Each resource usually has a stable **composition resource name** (e.g. annotation `krm.kcl.dev/composition-resource-name`). |
| **Desired composite resource** | Patches to the **XR** — commonly **status** updates (e.g. `clusterName`, `nodeGroupArn`). KCL merges into **`_dxr`**. |
| **Results** | Structured messages (severity, message) for **events** and debugging. |
| **Fatal** results | Can stop the pipeline and surface an error. |

Intuition: **“Apply these desired MRs; update the composite status like this.”**

---

## Pipeline flow (multiple functions)

```
RunFunction(step 1) → merge desired state
       ↓
RunFunction(step 2) → merge desired state
       ↓
Crossplane applies desired MRs to the API server
       ↓
Providers reconcile MRs to AWS / cloud
       ↓
Next reconcile: observed state includes new statuses → functions run again
```

Steps see **observed** state that may already reflect **earlier** steps in the same reconciliation (depending on runtime batching). Phased KCL (e.g. **`ready()`** / **`exists()`**) depends on **observed** composed resources from **previous** reconciles or earlier pipeline steps.

---

## Why this model exists

| Benefit | Explanation |
|--------|----------------|
| **Same contract for all runtimes** | **KCL**, **CUE** ([function-cue](https://github.com/crossplane-contrib/function-cue)), **Go**, **Python**, etc. — only the **implementation** of “compute response from request” changes. |
| **Separation of concerns** | Functions **emit desired Kubernetes objects**; **providers** enforce cloud state. |
| **Testability** | Inputs/outputs can be **serialized** (e.g. **`crossplane render`** feeds similar data without a full cloud). |

---

## Relation to this repo

| Doc | Connection |
|-----|------------|
| [CONFIGURATION-PIPELINE.md](./CONFIGURATION-PIPELINE.md) | XR → Composition → Function → image → **`main.k`** |
| [EKS-MAIN-K-EXPLAINED.md](./EKS-MAIN-K-EXPLAINED.md) | How **`functions/eks/main.k`** builds **`items`** from **`option("params")`** (observed XR + observed composed). Includes **“Is KCL just generating YAML?”** — desired K8s specs via code, not a file on disk. |
| [guides/15_Functions_Deep_Dive.md](../guides/15_Functions_Deep_Dive.md) | Go `RunFunction` signature and implementation details. |

---

## See also

- [Composition functions — Crossplane docs](https://docs.crossplane.io/latest/composition/compositions/)
- [function-sdk-go](https://github.com/crossplane/function-sdk-go) — Go types for `RunFunctionRequest` / `RunFunctionResponse`
