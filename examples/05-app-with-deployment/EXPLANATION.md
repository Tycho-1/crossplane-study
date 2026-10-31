# Example 05: XRD, Composition, XR, CRD, and Functions – Explained

This document explains **what an XRD is**, how it **relates to Composition and XR**, **how a CRD is created**, and **how a Composition Function works**, using this example (XApp → Deployment + Service).

---

## 1. What is an XRD?

An **XRD** (Composite Resource Definition) is a Crossplane resource that **defines a custom API**. It does not create any workload by itself; it tells the cluster: “there is a new kind of resource (e.g. `XApp`) with this schema.”

In this example, **xrd.yaml** defines:

| XRD field | In this example | Meaning |
|-----------|------------------|---------|
| **metadata.name** | `xapps.example.org` | Name of the XRD (often `plural.group`) |
| **spec.scope** | `Namespaced` | XApp instances live in a namespace (v2 default) |
| **spec.group** | `example.org` | API group of the custom resource |
| **spec.names.kind** | `XApp` | Kind users will create (e.g. `kind: XApp`) |
| **spec.names.plural** | `xapps` | Plural for the API URL (e.g. `/apis/example.org/v1alpha1/xapps`) |
| **spec.versions** | `v1alpha1` with schema | One or more versions; each has an **openAPIV3Schema** that defines `spec` (and optionally `status`) |

The **schema** in the XRD defines what users can set when they create an XApp:

```yaml
spec:
  image: string   # default: nginx
  replicas: int   # default: 1
required: [image]
```

So the XRD is the **contract**: “An XApp has `spec.image` and `spec.replicas`.”

---

## 2. How a CRD is created from the XRD

When you **apply the XRD**, Crossplane **creates a Kubernetes CRD** (Custom Resource Definition) that implements that API.

- **XRD** = the Crossplane object you write (e.g. `CompositeResourceDefinition` with name `xapps.example.org`).
- **CRD** = the resulting Kubernetes `CustomResourceDefinition` that registers the new resource type in the API server.

Flow:

1. You run: `kubectl apply -f xrd.yaml`
2. Crossplane’s controller sees the XRD.
3. Crossplane **creates or updates** a CRD whose:
   - **Group** = `example.org` (from `spec.group`)
   - **Version(s)** = `v1alpha1` (from `spec.versions`)
   - **Kind** = `XApp`, **Plural** = `xapps` (from `spec.names`)
   - **Scope** = Namespaced (from `spec.scope`)
   - **Schema** = the same `openAPIV3Schema` from the XRD

After that, the Kubernetes API serves the new type. You can do:

```bash
kubectl get xapps
kubectl get xapp my-app -o yaml
```

So: **XRD → Crossplane creates/updates CRD → API has XApp**. The CRD is the “real” API definition; the XRD is Crossplane’s way to define it and to link it to Compositions.

---

## 3. How XRD relates to Composition and XR

| Resource | Role in one sentence |
|----------|----------------------|
| **XRD** | Defines the **API** (XApp with spec.image, spec.replicas) and causes the **CRD** to exist. |
| **Composition** | Defines **how** that API is implemented: when someone creates an XApp, **run this pipeline** (e.g. function-patch-and-transform) to produce Deployment + Service. |
| **XR** | An **instance** of that API (e.g. `XApp` named `my-app` with `image: nginx`, `replicas: 2`). Creating/updating the XR triggers the Composition. |

Relationship:

- **XRD** says: “There is an API called XApp.”
- **Composition** says: “For the API XApp, when an XR is created, run this pipeline and create these composed resources (Deployment, Service).”
- **XR** is the actual object the user creates; Crossplane finds the right Composition (by `compositeTypeRef`) and runs the pipeline for that XR.

So: **XRD defines the API and CRD; Composition defines the implementation; XR is the instance that triggers the implementation.**

---

## 4. Flow in this example (step by step)

1. **Apply XRD** (`xrd.yaml`)  
   → Crossplane creates the CRD for `XApp` (group `example.org`, kind `XApp`).  
   → You can create XApp objects, but no Deployment/Service will be created yet.

2. **Apply Composition** (`composition.yaml`)  
   → Crossplane records: “For `compositeTypeRef` = XApp (example.org/v1alpha1), use this pipeline.”  
   → The pipeline has one step: call **function-patch-and-transform** with the given **input** (Resources with base Deployment, base Service, and patches).

3. **Apply XR** (`xr.yaml`)  
   → User creates one XApp instance: `metadata.name: my-app`, `spec.image: nginx`, `spec.replicas: 2`.  
   → Crossplane sees the XR, finds the Composition whose `compositeTypeRef` matches XApp.  
   → Crossplane runs the pipeline:  
     - Calls **function-patch-and-transform** with the Composition’s input and the **observed XR** (my-app, image=nginx, replicas=2).  
     - The function **applies patches** (e.g. FromCompositeFieldPath: spec.replicas → Deployment.spec.replicas, spec.image → container image, metadata.name → labels/selector).  
     - The function **outputs** the desired resources: one Deployment, one Service.  
   → Crossplane **applies** those resources to the cluster (in the same namespace as the XR, with ownership so they are tied to the XR).

4. **Result**  
   → You get a Deployment and a Service created and updated when you create/update the XApp XR.  
   → `kubectl get deployment,service -l crossplane.io/composite=my-app` shows them.

---

## 5. How the Function works (function-patch-and-transform)

The Composition does **not** contain raw Deployment/Service YAML that Crossplane applies directly. It contains a **function pipeline**. In this example there is **one step** that uses the **Patch & Transform** function.

### What is a Composition Function?

A **Composition Function** is a small program (running as a separate pod or in the same process) that:

- **Input:** Receives the **observed state** (the XR, and optionally existing composed resources) plus **function-specific input** from the Composition (e.g. a list of “resources” with bases and patches).
- **Output:** Returns the **desired** set of composed resources (and optionally readiness).

Crossplane calls the function, gets back the desired resources, then **applies** them to the cluster (creating/updating Deployment, Service, etc.).

### How Patch & Transform is used in this example

In **composition.yaml** you have:

```yaml
pipeline:
  - step: create-deployment-and-service
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
        - name: deployment
          base: { ... }   # base Deployment spec
          patches: [ ... ]  # patches from XR fields into the base
        - name: service
          base: { ... }
          patches: [ ... ]
```

- **functionRef.name** must match an installed **Function** (see **functions.yaml**). That tells Crossplane **which** function to run (the Patch & Transform image).
- **input** is the **Patch & Transform** format: a list of **resources**, each with:
  - **base:** A template resource (e.g. a Deployment or Service with default values).
  - **patches:** List of patch rules. **FromCompositeFieldPath** copies a value from the **XR** (e.g. `spec.replicas`, `spec.image`, `metadata.name`) into a path in the composed resource (e.g. `spec.replicas`, `spec.template.spec.containers[0].image`, `spec.selector.matchLabels[app]`).

When the function runs:

1. It receives the **XR** (e.g. my-app with image=nginx, replicas=2).
2. For each entry in `resources`, it starts from **base**, then applies each **patch** (e.g. set Deployment.spec.replicas from XR.spec.replicas, set container image from XR.spec.image, set labels/selector from XR.metadata.name).
3. It returns the **patched** Deployment and Service as the desired composed resources.
4. Crossplane applies those to the cluster (with the correct namespace and ownership).

So the **function** is what turns “XR + template + patches” into “concrete Deployment and Service YAML”. Without the function, Crossplane would not know how to generate those resources from the XR.

---

**Where is the code for the function?**  
Yes, there is source code. The function we use is **function-patch-and-transform**; its code lives in the repo [crossplane-contrib/function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform). The package (image) is built from that repo.

**What other functions exist?**  
You can use other composition functions in your pipeline, for example: **function-cue** (CUE scripts, loops/conditionals), **function-auto-ready** (readiness), or **custom functions** in **Go** (function-sdk-go, function-template-go) or **Python**. See the guide **[Composition Functions: Available and Custom](../../guides/Composition_Functions_Available_And_Custom.md)** for a list, examples, and how to write your own (Go: `crossplane xpkg init`, implement `RunFunction` in `fn.go`, build and push package).

---

## 6. Adding the Function resource (functions.yaml)

The Composition **references** a function by name: `functionRef.name: function-patch-and-transform`. That name must refer to an installed **Function** resource in the cluster.

In this example we add **functions.yaml**:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.8.2
```

- **Function** is a Crossplane package type (like Provider). Installing it makes the **function-patch-and-transform** controller available.
- When the Composition runs a step with `functionRef.name: function-patch-and-transform`, Crossplane invokes **this** function with the step’s **input** and the current XR state.
- The function’s **package** is the OCI image that implements Patch & Transform (base + patches → composed resources).

Apply order: **XRD → Function → Composition → XR**. So the CRD exists, the function is installed, the Composition is registered, and then creating the XR triggers the pipeline.

---

## 7. Summary diagram (this example)

```
XRD (xrd.yaml)          →  CRD created for XApp (example.org/v1alpha1)
Function (functions.yaml) →  function-patch-and-transform available
Composition (composition.yaml) →  "For XApp, run P&T with this base + patches"
XR (xr.yaml)            →  my-app (image=nginx, replicas=2)
                                ↓
                    Crossplane runs pipeline:
                      P&T function( XR + input ) → Deployment + Service
                                ↓
                    Crossplane applies Deployment + Service
                                ↓
                    kubectl get deployment,service -l crossplane.io/composite=my-app
```

---

## 8. Using another function (e.g. CUE, or a second P&T step)

You can add **more steps** to the pipeline. Each step calls one function and can receive different input.

**In this example**, the Composition has **two steps**:

1. **create-deployment-and-service** – P&T with input: Deployment base + patches, Service base + patches. Output: one Deployment, one Service.
2. **create-configmap** – P&T with input: ConfigMap base + patch (metadata.name from XR). Output: one ConfigMap.

Crossplane runs the pipeline in order; the first step’s output (plus the XR) is available to the second step. The final desired state is the union of all composed resources; Crossplane applies Deployment, Service, and ConfigMap.

Conceptually, adding another step looks like this (same function, different input):

```yaml
pipeline:
  - step: create-deployment-and-service
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources: [ ... ]  # Deployment + Service as now
  - step: create-configmap
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
        - name: configmap
          base:
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: app-config
            data:
              app: "true"
          patches:
            - type: FromCompositeFieldPath
              fromFieldPath: metadata.name
              toFieldPath: metadata.name
```

The first step outputs Deployment + Service; the second step receives that plus the XR and outputs a ConfigMap. Crossplane then applies **all** composed resources (Deployment, Service, ConfigMap). So **functions** are the building blocks of the pipeline; you can chain several (P&T, CUE, Python, etc.) as needed.
