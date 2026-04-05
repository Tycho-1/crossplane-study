# Up CLI DevEx Workflow Examples

Practical examples for using the **`up` CLI** (Upbound) to develop Crossplane XRDs, Compositions, and Functions efficiently.

## Prerequisites

```bash
# Install Up CLI
curl -sL "https://cli.upbound.io" | sh

# Verify
up --version
```

Docker is required for `up project build` and `up uxp install`.

---

## Example 1: Create a new project from scratch

```bash
# Initialize project
up project init my-s3-platform
cd my-s3-platform

# Add dependencies
up dependency add "xpkg.upbound.io/upbound/provider-aws-s3@>=v1.22.0"
up dependency add "xpkg.upbound.io/crossplane-contrib/function-auto-ready@>=v0.0.0"

# Inspect structure
ls -la apis/ examples/ functions/
cat upbound.yaml
```

---

## Example 2: Generate XRD from an example (inverted workflow)

**Idea:** Define what users will write first, then generate the XRD.

```bash
# Generate example claim
up example generate --name my-bucket \
    --namespace default \
    --type xrc \
    --kind StorageBucket \
    --api-group platform.example.com \
    --api-version v1alpha1

# Edit examples/storagebucket/my-bucket.yaml with desired spec, e.g.:
# spec:
#   parameters:
#     region: us-east-1
#     versioning: true

# Generate XRD from the example
up xrd generate examples/storagebucket/my-bucket.yaml

# Result: apis/xstoragebuckets/definition.yaml
```

---

## Example 3: Generate Composition boilerplate

```bash
# After you have an XRD at apis/xstoragebuckets/definition.yaml
up composition generate apis/xstoragebuckets/definition.yaml \
    --path xstoragebuckets/aws-s3.yaml

# Result: apis/xstoragebuckets/aws-s3.yaml with pipeline skeleton
```

---

## Example 4: Generate KCL function with schema support

```bash
# Scaffold a KCL function for the Composition
up function generate aws-s3 apis/xstoragebuckets/aws-s3.yaml

# Result: functions/aws-s3/main.k with:
# - imports of provider schemas (IntelliSense)
# - oxr, ocds, dcds variables
# - empty _items array

# Edit functions/aws-s3/main.k — add resources with VSCode autocomplete
```

---

## Example 5: Build and push

```bash
up project build
up project push --tag v0.1.0
```

Requires registry authentication (e.g. `up login` or `docker login`).

---

## Example 6: Install UXP (Upbound Crossplane) in cluster

```bash
# Install UXP into current cluster
up uxp install

# Or specific version
up uxp install 1.8.1-up.1
```

---

## Your existing project

Your `test-up-command/test-up` project was created with `up project init` and already has:

- `apis/xstoragebucket/definition.yaml` — XRD
- `apis/xstoragebucket/composition.yaml` — Composition
- `examples/xstoragebuckets/example.yaml` — Example XR
- `functions/compose-bucket/` — Go function (not KCL)

To extend it:

```bash
cd ../test-up-command/test-up

# Add another dependency
up dependency add "xpkg.upbound.io/crossplane-contrib/function-patch-and-transform@>=v0.10.0"

# Build
up project build
```

---

## See also

- [guides/17_Up_CLI_And_DevEx.md](../../guides/17_Up_CLI_And_DevEx.md) — full guide
- [Upbound Control Plane Projects](https://docs.upbound.io/core-concepts/projects/)
- [test-up-command/test-up](../../test-up-command/test-up/) — your project
