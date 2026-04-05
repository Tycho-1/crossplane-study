# Composition Functions — Deep Dive

Everything you need to know about how functions work, which ones to use, how to combine them, and how to write your own.

---

## What is a function?

A composition **function** is the executable step in a Composition pipeline. It receives the current state of the XR and any already-composed resources, applies logic, and returns the desired state of all composed resources.

```
User creates XR
      │
      ▼
Crossplane picks the matching Composition
      │
      ▼
┌─────────────────────────────────┐
│  Pipeline                       │
│  Step 1: function-kcl           │  ← creates infrastructure resources
│  Step 2: function-patch-and-transform │  ← adds patches / overrides
│  Step 3: function-auto-ready    │  ← derives XR readiness
└─────────────────────────────────┘
      │
      ▼
Crossplane applies the desired composed resources to the cluster
```

Each step gets:
- **`oxr`** — the observed XR (what the user wrote)
- **`ocds`** — the observed composed resources (what already exists in the cluster)
- **Input** — optional static config from the Composition YAML

Each step returns:
- **Desired composed resources** — the full list of resources Crossplane should create/update
- **Status patches** — optional updates to write back to the XR's `status`
- **Conditions/events** — optional health signals

Steps run **sequentially**. Each step can read the output of previous steps via `ocds`.

---

## How a function is installed

A function is just a Crossplane package (OCI image) installed with a `Function` CR:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-patch-and-transform:v0.10.1
```

```bash
kubectl get functions.pkg.crossplane.io

NAME                           INSTALLED   HEALTHY   AGE
function-patch-and-transform   True        True      5m
```

The function runs as a Pod in `crossplane-system`. Crossplane calls it via gRPC for every reconcile of every XR that uses it.

---

## All available functions

### function-patch-and-transform (most common)

The standard function. Declarative YAML patches — no real programming logic but covers most simple cases.

**Install:**
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-patch-and-transform:v0.10.1
```

**Use in Composition:**
```yaml
- step: patch-resources
  functionRef:
    name: function-patch-and-transform
  input:
    apiVersion: pt.fn.crossplane.io/v1beta1
    kind: Resources
    resources:
      - name: bucket
        base:
          apiVersion: s3.aws.m.upbound.io/v1beta1
          kind: Bucket
          spec:
            forProvider:
              region: us-east-2
        patches:
          - type: FromCompositeFieldPath
            fromFieldPath: spec.region
            toFieldPath: spec.forProvider.region
```

**When to use:** Simple 1-to-1 field mapping, static resource templates, no loops or conditionals needed.

**Limitation:** No loops, no conditionals, no cross-resource value injection (e.g. you can't inject a bucket ARN into an IAM policy without a more powerful function).

---

### function-kcl

Uses [KCL](https://kcl-lang.io/) — a Python-like configuration language. Supports full programming logic: conditionals, loops, computed values, cross-resource references.

**Install:**
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-kcl
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-kcl:v0.10.3
```

**Use in Composition:**
```yaml
- step: create-resources
  functionRef:
    name: function-kcl
  input:
    apiVersion: krm.kcl.dev/v1alpha1
    kind: KCLRun
    spec:
      source: |
        oxr = option("params").oxr

        # Map a human-readable size to a machine type
        _machine_type = {
          "small":  "t3.medium",
          "medium": "t3.large",
          "large":  "t3.xlarge",
        }[oxr.spec.parameters.nodeSize]

        # Conditionally create a GPU node group
        _items = [{
          apiVersion = "ec2.aws.m.upbound.io/v1beta1"
          kind = "Instance"
          metadata.name = oxr.metadata.name
          spec.forProvider = {
            instanceType = _machine_type
            region = oxr.spec.region
          }
        }]

        if oxr.spec.parameters?.gpu?.enabled:
          _items += [{
            apiVersion = "ec2.aws.m.upbound.io/v1beta1"
            kind = "Instance"
            metadata.name = oxr.metadata.name + "-gpu"
            spec.forProvider = {
              instanceType = "p3.2xlarge"
              region = oxr.spec.region
            }
          }]

        items = _items
```

**When to use:** Any time you need conditionals, loops, or computed values. Preferred for complex compositions (like `dot-kubernetes` which uses KCL to generate EKS/GKE/AKS clusters). Replaces `function-patch-and-transform` for non-trivial cases.

**Key variables:**
- `oxr` — the observed XR (`option("params").oxr`)
- `ocds` — the observed composed resources (`option("params").ocds`)
- `items` — what you return (the list of desired composed resources)

---

### function-go-templating

Uses Go's `text/template` syntax — familiar if you know Helm. Good for teams comfortable with Helm templating.

**Install:**
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-go-templating
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-go-templating:v0.7.0
```

**Use in Composition:**
```yaml
- step: create-bucket
  functionRef:
    name: function-go-templating
  input:
    apiVersion: gotemplating.fn.crossplane.io/v1beta1
    kind: GoTemplate
    source: Inline
    inline:
      template: |
        apiVersion: s3.aws.m.upbound.io/v1beta1
        kind: Bucket
        metadata:
          name: {{ .observed.composite.resource.metadata.name }}
          annotations:
            crossplane.io/external-name: {{ .observed.composite.resource.spec.storageName }}
        spec:
          forProvider:
            region: {{ .observed.composite.resource.spec.region }}
        ---
        {{ if .observed.composite.resource.spec.parameters.enableVersioning }}
        apiVersion: s3.aws.m.upbound.io/v1beta1
        kind: BucketVersioning
        metadata:
          name: {{ .observed.composite.resource.metadata.name }}-versioning
        spec:
          forProvider:
            bucketRef:
              name: {{ .observed.composite.resource.metadata.name }}
            versioningConfiguration:
              - status: Enabled
        {{ end }}
```

**When to use:** Teams migrating from Helm who prefer `{{ }}` syntax over KCL. Good for YAML-heavy compositions with conditionals.

---

### function-auto-ready

Derives the XR's readiness from the readiness of its composed resources. Without this, an XR reports `Ready: True` immediately even if the S3 bucket hasn't been created yet.

**Install:**
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-auto-ready
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-auto-ready:v0.2.1
```

**Use in Composition — always as the last step:**
```yaml
pipeline:
  - step: create-resources
    functionRef:
      name: function-patch-and-transform
    input: ...

  - step: detect-readiness       # always last
    functionRef:
      name: function-auto-ready
```

**When to use:** Almost always — add it as the final step in every Composition pipeline so your XR accurately reflects when resources are actually ready.

---

### function-sequencer

Controls the order in which composed resources are created. Useful when resource B depends on resource A being ready before it can be created (e.g. an RDS instance must exist before a database user).

**Install:**
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-sequencer
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-sequencer:v0.3.0
```

**Use in Composition:**
```yaml
- step: sequence
  functionRef:
    name: function-sequencer
  input:
    apiVersion: sequencer.fn.crossplane.io/v1beta1
    kind: Input
    rules:
      - sequence:
          - rds-instance     # must be ready before...
          - rds-db-user      # ...this is created
```

---

### Choosing the right function

| Scenario | Recommended function |
|---|---|
| Simple field mapping, static templates | `function-patch-and-transform` |
| Conditionals, loops, computed values | `function-kcl` |
| Helm-style templating | `function-go-templating` |
| XR readiness from composed resources | `function-auto-ready` (always add as last step) |
| Resource creation ordering | `function-sequencer` |
| Custom business logic / API calls | Custom function (Go or Python) |

---

## Combining functions in a pipeline

Functions are most powerful when combined. A typical production pipeline:

```yaml
pipeline:
  # Step 1: KCL generates all cloud resources with conditional logic
  - step: create-infrastructure
    functionRef:
      name: function-kcl
    input:
      apiVersion: krm.kcl.dev/v1alpha1
      kind: KCLRun
      spec:
        source: |
          ...

  # Step 2: patch-and-transform adds a few overrides on top
  - step: patch-overrides
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
        - name: bucket
          patches:
            - type: FromCompositeFieldPath
              fromFieldPath: spec.tags
              toFieldPath: spec.forProvider.tags

  # Step 3: auto-ready derives XR readiness (always last)
  - step: detect-readiness
    functionRef:
      name: function-auto-ready
```

---

## Writing a simple custom function (Go)

When no existing function covers your need, write your own. Here is a minimal working example that reads a field from the XR and creates a tagged S3 bucket.

### Prerequisites

```bash
# Go 1.21+
go version

# Docker (to build the function image)
docker version

# Crossplane CLI
crossplane --version
```

### Step 1 — Initialise from the template

```bash
crossplane xpkg init my-function function-template-go -d my-function
cd my-function
```

This creates:

```
my-function/
├── fn.go              ← implement your logic here
├── fn_test.go         ← unit tests
├── main.go            ← entrypoint (don't touch)
├── go.mod
├── Dockerfile
└── package/
    └── crossplane.yaml
```

### Step 2 — Implement the logic in fn.go

Replace the `RunFunction` body with your logic:

```go
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
    rsp := response.To(req, response.DefaultTTL)

    // Read the XR
    oxr, err := request.GetObservedCompositeResource(req)
    if err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot get observed XR"))
        return rsp, nil
    }

    // Read a field from spec
    region, err := oxr.Resource.GetString("spec.region")
    if err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot get spec.region"))
        return rsp, nil
    }

    storageName, err := oxr.Resource.GetString("spec.storageName")
    if err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot get spec.storageName"))
        return rsp, nil
    }

    // Build the desired composed resources
    desired, err := request.GetDesiredComposedResources(req)
    if err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
        return rsp, nil
    }

    // Create a Bucket resource
    bucket := composed.New()
    bucket.Resource.SetAPIVersion("s3.aws.m.upbound.io/v1beta1")
    bucket.Resource.SetKind("Bucket")
    if err := bucket.Resource.SetString("spec.forProvider.region", region); err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot set region"))
        return rsp, nil
    }
    bucket.Resource.SetName(storageName)

    desired["bucket"] = bucket

    if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
        response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
        return rsp, nil
    }

    return rsp, nil
}
```

### Step 3 — Test locally

```bash
# Run tests
go test -v -cover ./...

# Run the function locally (terminal 1)
go run . --insecure --debug

# Test with crossplane render (terminal 2)
# Add this annotation to your functions.yaml Function:
#   render.crossplane.io/runtime: Development
crossplane render xr.yaml composition.yaml functions.yaml
```

### Step 4 — Build and push

Unlike Configurations (which are pure YAML and don't need Docker), a Function package embeds a real controller binary. Building requires two steps: first `docker build` to compile the binary into a runtime image, then `crossplane xpkg build` to wrap it into a Crossplane package.

```bash
# Step 1: build the runtime image (compiles the Go binary)
docker build --platform=linux/amd64 -t my-function-runtime:latest .

# Step 2: build the Crossplane package (embeds the runtime image into an OCI package)
crossplane xpkg build \
  --package-root=package/ \
  --embed-runtime-image=my-function-runtime:latest \
  --output=my-function.xpkg

# Push — any OCI tool works here (crossplane xpkg push, docker push, crane, skopeo)
crossplane xpkg push \
  --package=my-function.xpkg \
  docker.io/yourorg/my-function:v0.1.0
```

> See [Packages & xpkg CLI](./14_Packages_And_xpkg_CLI.md) for the full breakdown of build vs push tooling and when Docker is required.

### Step 5 — Install and use in a Composition

```yaml
# Install the function
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: my-function
spec:
  package: docker.io/yourorg/my-function:v0.1.0
```

```yaml
# Use in Composition pipeline
- step: my-custom-step
  functionRef:
    name: my-function
```

---

## When to write a custom function vs use an existing one

| Use case | Solution |
|---|---|
| Field mapping / patching | `function-patch-and-transform` |
| Conditionals and loops | `function-kcl` |
| Call an external HTTP API during composition | Custom function (Go) |
| Generate resources from a database query | Custom function (Go/Python) |
| Complex multi-step business logic | Custom function (Go) |
| Proprietary internal naming conventions | Custom function (Go/Python) |

Rule of thumb: **always try an existing function first**. KCL covers most cases that `function-patch-and-transform` cannot. Only write a custom function when you need to call external systems or implement logic that no configuration language can express.

---

## See also

- [Composition Functions: Available and Custom](./Composition_Functions_Available_And_Custom.md) — original overview
- [Composition: XRDs, Compositions & Functions](./05_Composition_XRDs_Functions.md) — how functions fit into Compositions
- [Packages & xpkg CLI](./14_Packages_And_xpkg_CLI.md) — building and pushing function packages
- [function-patch-and-transform docs](https://docs.crossplane.io/latest/guides/function-patch-and-transform/)
- [Write a Composition Function in Go](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-go/)
- [Write a Composition Function in Python](https://docs.crossplane.io/latest/guides/write-a-composition-function-in-python/)
- [function-sdk-go](https://github.com/crossplane/function-sdk-go)
- [function-template-go](https://github.com/crossplane/function-template-go)
- [KCL language](https://kcl-lang.io/)
