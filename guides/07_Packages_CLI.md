# Packages & CLI

## Packages overview

Crossplane has three main **package** types:

| Type | Purpose | Example |
|------|---------|--------|
| **Provider** | MRs + controller for an external system | provider-aws-s3, provider-azure |
| **Function** | Composition function used in pipelines | function-patch-and-transform, function-cue |
| **Configuration** | Bundle of XRDs, Compositions, and optionally dependencies | Platform configs |

All use **fully qualified package URLs** in v2 (e.g. `xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0`).

---

## Installing packages

- **Provider:** create a `Provider` resource with `spec.package`.
- **Function:** create a `Function` resource with `spec.package`.
- **Configuration:** create a `Configuration` resource with `spec.package`.

Optional fields: `packagePullPolicy`, `revisionActivationPolicy`, `revisionHistoryLimit`, `packagePullSecrets`, `runtimeConfigRef` (for Provider/Function).

### Where to get packages: Upbound Marketplace

The **Upbound Marketplace** ([marketplace.upbound.io](https://marketplace.upbound.io)) is the main place to discover package names, versions, and install manifests for Providers, Configurations, and Functions. See **[Upbound Marketplace](./11_Upbound_Marketplace.md)** for why it matters, how to use it, and the difference between Providers and Configurations (e.g. [configuration-aws-eks](../configuration-aws-eks/) vs [eks-cluster-spot](../eks-cluster-spot/)).

---

## Crossplane CLI

The **crossplane** CLI helps you work with Compositions and functions **without a live cluster** (e.g. render composed resources locally).

### Install CLI

See [CLI Reference](https://docs.crossplane.io/latest/cli/). Typical options:

- Download from [releases](https://github.com/crossplane/crossplane/releases).
- Or use a package manager if available for your OS.

### crossplane render

**Render** takes an XR, the Compositions it could use, and the Functions referenced by those Compositions, and **outputs** the XR plus all composed resources as YAML. Uses Docker to run the function containers.

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

- **xr.yaml:** One composite resource (the “observed” input to the pipeline).
- **composition.yaml:** One or more Compositions (the one matching the XR’s kind is used).
- **functions.yaml:** The Function resources (by name) that the Composition references; CLI uses them to run the correct function images.

Output: multi-document YAML to stdout. You can diff or pipe to `kubectl apply -f -` (e.g. in CI).

**Development runtime:** For local function development, you can run the function yourself (e.g. `go run . --insecure`) and point the CLI at it with the `Development` runtime and `render.crossplane.io/runtime-development-target` annotation so it doesn’t pull/run Docker.

---

## Function annotations for render

On the **Function** resource you can use:

- `render.crossplane.io/runtime`: `Docker` (default) or `Development`.
- `render.crossplane.io/runtime-docker-cleanup`: `Stop` or `Orphan` (when using Docker).
- `render.crossplane.io/runtime-docker-pull-policy`: `Always`, `Never`, `IfNotPresent`.
- `render.crossplane.io/runtime-development-target`: gRPC target when using Development (e.g. `localhost:9443`).

---

## Package lifecycle

- **Upgrade:** Change `spec.package` to a new version and apply; Crossplane creates a new revision and (by default) activates it.
- **Revisions:** `kubectl get providerrevisions`, `kubectl get functionrevisions` (and similar for Configuration). One revision is **Active**; others are **Inactive** (kept for history if `revisionHistoryLimit` > 0).
- **Remove:** Delete the Provider/Function/Configuration. Remove or orphan dependent resources first to avoid abandoned MRs or broken Compositions.

---

## Image configs and signature verification

Crossplane can verify package images via **ImageConfig** and optional **signature verification** (alpha, `--enable-signature-verification`). See [Image Configs](https://docs.crossplane.io/latest/packages/image-configs/) and [Installation](./03_Installation_Setup.md) for feature flags.

---

## Next

- [Upgrade & Troubleshooting](./08_Upgrade_Troubleshooting.md).
- Try `crossplane render` with [examples/04-patch-and-transform](../examples/04-patch-and-transform/).
