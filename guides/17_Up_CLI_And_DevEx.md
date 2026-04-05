# Up CLI and DevEx for Crossplane

A guide to the **`up` CLI** (Upbound) and how it improves the developer experience for building XRDs, Compositions, and Functions — based on [Save Hours with DevEx for Crossplane](https://devopstoolkit.live/crossplane/save-hours-with-devex-for-crossplane/index.html) by Viktor Farcic.

---

## Up CLI vs Crossplane CLI — they are different

| | **`up`** (Upbound CLI) | **`crossplane`** (Crossplane CLI) |
|---|---|---|
| **Source** | [Upbound](https://github.com/upbound/up) | [Crossplane](https://github.com/crossplane/crossplane) |
| **Docs** | [docs.upbound.io](https://docs.upbound.io/manuals/cli/overview) | [docs.crossplane.io/cli](https://docs.crossplane.io/v2.2/cli/) |
| **Purpose** | Project scaffolding, XRD/Composition generation, build & push | Package build, render, validate |
| **Typical use** | Day-to-day development of Configurations | One-off builds, local render, troubleshooting |

**Why `up` is not in Crossplane docs:** The `up` CLI is an **Upbound** tool. Crossplane docs only cover the **Crossplane** CLI. Upbound is the commercial entity behind Crossplane; they ship `up` for their control plane projects and UXP (Upbound Crossplane). You need both CLIs for different workflows.

---

## Install the Up CLI

```bash
curl -sL "https://cli.upbound.io" | sh
```

Or with a specific version:

```bash
curl -sL "https://cli.upbound.io" | VERSION=v0.44.3 sh
```

See [Upbound CLI docs](https://docs.upbound.io/manuals/cli/overview) for Homebrew, Linux packages, and Windows.

---

## What the Up CLI gives you

### 1. Project scaffolding

Instead of manually creating folders and files, you get a standard layout:

```bash
up project init my-platform
cd my-platform
```

Creates:

```
my-platform/
├── apis/           ← XRDs and Compositions
├── examples/       ← Example claims/XR instances (user experience)
├── functions/      ← KCL or Python composition functions
└── upbound.yaml    ← Project metadata + dependsOn (like crossplane.yaml)
```

### 2. XRD from examples (inverted workflow)

**Traditional:** Write the XRD first, then figure out what users will write.

**DevEx:** Define the **example** (what users will write) first, then generate the XRD from it.

```bash
# Create an example claim/XR
up example generate --name my-db --namespace default --type xrc \
    --kind SQL --api-group devopstoolkit.live --api-version v1alpha1

# Edit examples/sql/my-db.yaml with the desired spec
# Then generate the XRD from it:
up xrd generate examples/sql/my-db.yaml
```

Output: `apis/xsqls/definition.yaml` — a full XRD derived from your example.

### 3. Composition boilerplate

Generate the Composition skeleton (pipeline structure, function refs) without writing YAML by hand:

```bash
up composition generate apis/xsqls/definition.yaml --path xsqls/aws.yaml
```

### 4. Function scaffolding with IntelliSense

Generate a KCL (or Python) function with schema imports for auto-complete in VSCode:

```bash
up function generate aws apis/xsqls/aws.yaml
```

Creates `functions/aws/main.k` with:
- Imports of provider schemas (from your `dependsOn`)
- Variables for `oxr`, `ocds`, `dcds`
- Empty `_items` array to fill with managed resources

VSCode + KCL extension gives **IntelliSense** — type `v1beta1.` and get autocomplete for all EC2/RDS resources.

### 5. Dependency management

Add providers and functions declaratively:

```bash
up dependency add "xpkg.upbound.io/upbound/provider-aws-s3@>=v1.22.0"
up dependency add "xpkg.upbound.io/crossplane-contrib/function-auto-ready@>=v0.0.0"
```

Updates `upbound.yaml` with `dependsOn`. Schemas are downloaded for IntelliSense.

### 6. Build and push

```bash
up project build
up project push --tag v0.0.1
```

Builds the full Configuration package and pushes to your registry.

---

## Configuration = Terraform module

An `up` project is the **source** of a Configuration. After `up project build` and `up project push`, you get a Configuration package — the Crossplane equivalent of a Terraform module.

| Terraform | Crossplane |
|-----------|------------|
| **Module** (variables.tf + main.tf in a folder) | **Configuration** (XRD + Compositions in a project) |
| `terraform init` + `terraform apply` | `kubectl apply` a Configuration CR |
| Module source (Git, registry) | OCI image (e.g. `xpkg.upbound.io/org/name:v1.0.0`) |
| `required_providers` | `dependsOn` in upbound.yaml / crossplane.yaml |

- **Terraform module** = folder with `.tf` files → published to registry
- **Crossplane Configuration** = `up` project (or `my-platform/` with crossplane.yaml) → built and pushed as OCI image

Both are reusable, versioned units that define an API and its implementation.

---

## Landing zone: one Configuration per capability

For a landing zone, you model each capability as its own Configuration (module):

| Landing zone capability | Configuration (module) | XR kind |
|-------------------------|-------------------------|---------|
| Networking | `configuration-aws-network` | `XNetwork`, `XVPC` |
| Object storage | `configuration-aws-s3` | `XBucket`, `XStorageBucket` |
| Databases | `configuration-aws-rds` | `XDatabase`, `XPostgreSQL` |
| Compute / clusters | `configuration-aws-eks` | `XCluster`, `XEKS` |
| IAM / security | `configuration-aws-iam` | `XRole`, `XPolicy` |
| **Applications** (v2) | `configuration-xapp` | `XApp` → Deployment + Service + ConfigMap |

**Flow:**

```
1. up project init configuration-aws-s3
2. up dependency add provider-aws-s3, function-patch-and-transform, ...
3. Define examples (what users request)
4. up xrd generate → XRD
5. up composition generate → Composition
6. up function generate → KCL/Go function
7. Implement the function (S3 bucket + optional policies, etc.)
8. up project build && up project push --tag v1.0.0
```

Each project becomes a separate Configuration package. App teams install them and create XR instances (claims) to provision resources.

**One vs many Configurations:**

- **One Configuration per capability** — e.g. `configuration-aws-s3`, `configuration-aws-rds`. Easier to version and adopt independently.
- **One Configuration for the whole landing zone** — e.g. `configuration-landing-zone` with many XRDs. Simpler to install, but less flexible.

Most teams use one Configuration per capability and treat them as modules in the landing zone.

**Application Configurations (v2):** Compositions can also create **Kubernetes applications** (Deployment, Service, ConfigMap) — not only cloud MRs. An `XApp` that deploys an app can be packaged as a Configuration the same way. See [Configurations: Application compositions](./13_Configurations.md#application-compositions-v2-as-configurations) and [examples/05-app-with-deployment](../examples/05-app-with-deployment/).

---

## Use cases

| Use case | Command(s) |
|----------|------------|
| Start a new Crossplane project | `up project init <name>` |
| Add a provider or function | `up dependency add <package>` |
| Define user experience first, then XRD | `up example generate` → edit → `up xrd generate` |
| Create Composition skeleton | `up composition generate <xrd-path> --path <output>` |
| Scaffold KCL/Python function with schemas | `up function generate <name> <composition-path>` |
| Build and publish | `up project build` → `up project push --tag <version>` |
| Install UXP in cluster | `up uxp install` |

---

## Workflow summary (DevEx style)

```
1. up project init my-platform
2. up dependency add <providers and functions>
3. up example generate ...  (or create examples/ manually)
4. Edit examples/*.yaml — define the ideal user experience
5. up xrd generate examples/...  → XRD created from examples
6. up composition generate apis/.../definition.yaml --path ...
7. up function generate <name> apis/.../composition.yaml
8. Edit functions/<name>/main.k — add managed resources with IntelliSense
9. up project build && up project push --tag v0.0.1
```

---

## Your existing project: test-up-command/test-up

You already ran `up project init` and have a working structure:

```
test-up/
├── apis/xstoragebucket/
│   ├── definition.yaml   ← XRD (XStorageBucket)
│   └── composition.yaml  ← Composition
├── examples/xstoragebuckets/example.yaml  ← Example XR
├── functions/compose-bucket/  ← Go function (not KCL)
└── upbound.yaml
```

This is the standard layout. Next steps you can try:

```bash
cd test-up-command/test-up

# Add another dependency
up dependency add "xpkg.upbound.io/crossplane-contrib/function-patch-and-transform@>=v0.10.0"

# Generate a new example and XRD (if you want a new API)
up example generate --name my-bucket --namespace default --type xr \
    --kind StorageBucket --api-group platform.example.com --api-version v1alpha1

# Build and push (requires registry auth)
up project build
up project push --tag v0.1.0
```

---

## Requirements (no Devbox needed)

- **Docker** — for building packages and running UXP locally
- **`up` CLI** — `curl -sL "https://cli.upbound.io" | sh`
- **Optional:** VSCode + KCL extension for IntelliSense when using KCL functions

Devbox is optional — it just bundles these tools. You can install them directly.

---

## Quick reference: up vs crossplane

| Task | Use |
|------|-----|
| Create project, add deps, generate XRD/Composition/Function | `up` |
| Build Configuration package | `up project build` or `crossplane xpkg build` |
| Push package to registry | `up project push` or `crossplane xpkg push` |
| Render Composition locally (dry-run) | `crossplane render` |
| Install UXP in cluster | `up uxp install` |

---

## Why the up CLI is not in Crossplane docs

The [Crossplane CLI Reference](https://docs.crossplane.io/v2.2/cli/) documents only the **`crossplane`** (or `crank`) binary — the official Crossplane CLI. It covers:

- `crossplane xpkg build` / `crossplane xpkg push`
- `crossplane render`
- `crossplane beta validate` / `crossplane beta trace`

The **`up`** CLI is **Upbound's** tool. It lives in [Upbound's documentation](https://docs.upbound.io/manuals/cli/overview). Upbound is the commercial entity behind Crossplane; they maintain UXP (Upbound Crossplane) and the `up` CLI for project scaffolding, XRD generation from examples, and build/push workflows. For efficient development of XRDs, Compositions, and Functions, you need both: `up` for the DevEx workflow, `crossplane` for render and low-level package operations.

---

## References

- [Save Hours with DevEx for Crossplane](https://devopstoolkit.live/crossplane/save-hours-with-devex-for-crossplane/index.html) — original article (Viktor Farcic)
- [Upbound CLI Overview](https://docs.upbound.io/manuals/cli/overview)
- [Control Plane Projects](https://docs.upbound.io/core-concepts/projects/)
- [Crossplane CLI Reference](https://docs.crossplane.io/v2.2/cli/) — `crossplane` (not `up`)
- [Crossplane Documentation v2.2](https://docs.crossplane.io/v2.2/)
- [Up CLI GitHub](https://github.com/upbound/up)
- [examples/up-cli-devex](../examples/up-cli-devex/) — workflow examples
