# Crossplane Packages & the xpkg CLI

A deep dive into what packages are, how they work, and how to build, push, and manage them with the Crossplane CLI.

---

## What is a package?

A Crossplane **package** is a versioned **OCI image** (the same format as a Docker image) that Crossplane pulls and installs into a cluster. Every Provider, Function, and Configuration is a package.

Under the hood a package image contains:
- One or more YAML manifests (XRDs, Compositions, CRDs, RBAC rules)
- A `crossplane.yaml` metadata file declaring the package type and dependencies
- Optionally: a controller binary (for Providers and Functions)

Crossplane pulls the image, extracts the YAML, applies it to the cluster, and — for Providers/Functions — runs the embedded controller as a Pod.

---

## The three package types

| Type | Kind to install | What it contains | Auto-installs |
|---|---|---|---|
| **Provider** | `Provider` | CRDs (MR schemas) + provider controller binary | nothing (see [MRDs](#mrd-managed-resource-definitions) for selective activation) |
| **Function** | `Function` | Function controller binary | nothing |
| **Configuration** | `Configuration` | XRDs + Compositions + `crossplane.yaml` | declared `dependsOn` packages |

### Provider

Installs the CRDs for a set of managed resources and runs the controller that reconciles them against the real cloud API.

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws-s3
spec:
  package: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0
```

```bash
# Check status
kubectl get providers.pkg.crossplane.io

NAME               INSTALLED   HEALTHY   PACKAGE                                                        AGE
provider-aws-s3    True        True      xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0   5m
```

Once healthy, the `Bucket`, `BucketPolicy` etc. CRDs are available in the cluster.

### Function

Installs a composition function that can be referenced in a Composition pipeline step.

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.10.1
```

```bash
kubectl get functions.pkg.crossplane.io

NAME                           INSTALLED   HEALTHY   PACKAGE                                                                          AGE
function-patch-and-transform   True        True      xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.10.1          3m
```

### Configuration

Installs XRDs and Compositions, and auto-installs all declared Provider/Function dependencies.

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
spec:
  package: docker.io/yourorg/my-platform:v0.1.0
```

```bash
kubectl get configurations.pkg.crossplane.io

NAME          INSTALLED   HEALTHY   PACKAGE                              AGE
my-platform   True        True      docker.io/yourorg/my-platform:v0.1.0 2m
```

---

## Package registries

Packages are stored in any OCI-compatible registry. Common choices:

| Registry | URL pattern | Use case |
|---|---|---|
| Upbound Marketplace | `xpkg.upbound.io/<account>/<name>:<tag>` | Official + community packages |
| Crossplane contrib | `xpkg.crossplane.io/crossplane-contrib/<name>:<tag>` | Community providers/functions |
| Docker Hub | `docker.io/<org>/<name>:<tag>` | Internal packages (public) |
| GitHub Container Registry | `ghcr.io/<org>/<name>:<tag>` | Internal packages (private) |
| AWS ECR | `<account>.dkr.ecr.<region>.amazonaws.com/<name>:<tag>` | Enterprise internal |

For private registries, create a pull secret and reference it:

```yaml
spec:
  package: ghcr.io/mycompany/my-platform:v1.0.0
  packagePullSecrets:
    - name: ghcr-pull-secret    # a kubernetes.io/dockerconfigjson Secret
```

---

## Package revisions

Every time you update a package version, Crossplane creates a new **revision** rather than mutating in place. This gives you a full upgrade history and the ability to roll back.

```bash
# See all revisions for a provider
kubectl get providerrevisions.pkg.crossplane.io

NAME                        HEALTHY   REVISION   CURRENT   AGE
provider-aws-s3-abc123      True      1          False     10d   ← old, inactive
provider-aws-s3-def456      True      2          True      1d    ← active
```

```bash
# Same for functions and configurations
kubectl get functionrevisions.pkg.crossplane.io
kubectl get configurationrevisions.pkg.crossplane.io
```

**Revision activation policies:**

| Policy | Behaviour |
|---|---|
| `Automatic` (default) | New revision becomes active immediately on install/upgrade |
| `Manual` | New revision stays inactive; you activate it manually |

Use `Manual` in production when you want to stage upgrades:

```yaml
spec:
  package: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.1.0
  revisionActivationPolicy: Manual
```

---

## Upgrading a package

Change the `spec.package` field and re-apply. Crossplane pulls the new image and creates a new revision:

```bash
# Edit the version in the manifest and apply
kubectl apply -f provider-aws-s3.yaml

# Or patch directly
kubectl patch provider.pkg.crossplane.io provider-aws-s3 \
  --type merge \
  -p '{"spec":{"package":"xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.1.0"}}'

# Watch the new revision become active
kubectl get providerrevisions.pkg.crossplane.io --watch
```

---

## The crossplane xpkg CLI

The `crossplane xpkg` subcommand is used to **build** and **push** Configuration (and Provider/Function) packages. It produces a valid OCI image from a folder of YAML files.

### Build

```bash
crossplane xpkg build \
  --package-root=my-platform/ \   # folder containing crossplane.yaml + XRDs + Compositions
  --output=my-platform.xpkg       # output file (.xpkg is just a renamed OCI image tarball)
```

What `xpkg build` does:
1. Reads `crossplane.yaml` from the package root
2. Collects all `.yaml` files in the folder (XRDs, Compositions)
3. Validates the YAML (basic schema check)
4. Packs everything into an OCI image tarball

**Useful flags:**

| Flag | Default | Description |
|---|---|---|
| `--package-root` | `.` | Directory containing `crossplane.yaml` |
| `--output` | `<name>.xpkg` | Output file path |
| `--ignore` | — | Glob patterns to exclude files (e.g. `examples/*`) |

```bash
# Exclude example files from the package
crossplane xpkg build \
  --package-root=. \
  --output=my-platform.xpkg \
  --ignore="examples/**" \
  --ignore="*.md"
```

### Push

```bash
crossplane xpkg push \
  --package=my-platform.xpkg \
  docker.io/yourorg/my-platform:v0.1.0
```

For private registries, make sure you are logged in first:

```bash
docker login ghcr.io
crossplane xpkg push \
  --package=my-platform.xpkg \
  ghcr.io/mycompany/my-platform:v0.1.0
```

### Build and push in one step

You can skip the intermediate `.xpkg` file and push directly:

```bash
crossplane xpkg build \
  --package-root=my-platform/ \
  --output=my-platform.xpkg && \
crossplane xpkg push \
  --package=my-platform.xpkg \
  docker.io/yourorg/my-platform:v0.1.0
```

### Typical CI/CD pipeline (GitHub Actions)

```yaml
# .github/workflows/release.yaml
on:
  push:
    tags: ["v*"]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Crossplane CLI
        run: |
          curl -sL https://raw.githubusercontent.com/crossplane/crossplane/main/install.sh | sh
          sudo mv crossplane /usr/local/bin/

      - name: Log in to registry
        run: echo "${{ secrets.GHCR_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Build package
        run: crossplane xpkg build --package-root=. --output=package.xpkg

      - name: Push package
        run: |
          crossplane xpkg push \
            --package=package.xpkg \
            ghcr.io/${{ github.repository }}:${{ github.ref_name }}
```

---

## Other useful CLI commands

### crossplane render

Dry-run a Composition pipeline locally without a cluster. Needs Docker to run function images.

```bash
crossplane render xr.yaml composition.yaml functions.yaml
```

Output is multi-document YAML to stdout — all the composed resources that would be created.

```bash
# Diff against current cluster state (useful in CI)
crossplane render xr.yaml composition.yaml functions.yaml | kubectl diff -f -
```

### crossplane validate (beta)

Validate XRDs, Compositions, and XR instances against schemas without applying them.

```bash
crossplane beta validate --schema-dir=. xr.yaml
```

### crossplane trace (beta)

Trace the full resource tree of an XR — shows the XR, all composed resources, and their status/conditions. Very useful for debugging.

```bash
crossplane beta trace xplatform my-platform -n default
```

Example output:
```
NAME                                               SYNCED   READY   STATUS
XPlatform/my-platform (example.org/v1alpha1)       True     True    Available
└─ Bucket/my-platform-storage                      True     True    Available
└─ Table/my-platform-lock-table                    True     True    Available
└─ Policy/my-platform-s3-policy                    True     True    Available
```

---

## What is a package file (.xpkg) exactly?

There are two different cases depending on the package type. Understanding this explains why the build tools differ.

### Configuration — no Dockerfile, no container runtime

A Configuration package contains **only YAML files** (XRDs, Compositions, `crossplane.yaml`) packed into an OCI image. There is no Dockerfile and no running process inside it. `crossplane xpkg build` assembles the OCI image layers directly from the YAML files — **Docker is not required** to build it.

```
my-platform/
├── crossplane.yaml    ← package metadata
├── xrd.yaml           ← packed as OCI layer
└── composition.yaml   ← packed as OCI layer

→ crossplane xpkg build → OCI image (no Docker needed)
```

### Function / Provider — has a Dockerfile and an embedded binary

These packages embed a **controller binary** (a running Go process) that Crossplane executes as a Pod. Building them requires two steps:

```
my-function/
├── fn.go              ← Go source code
├── Dockerfile         ← builds the binary
└── package/
    └── crossplane.yaml

→ Step 1: docker build → runtime image (the running process)
→ Step 2: crossplane xpkg build --embed-runtime-image → final OCI package
```

The resulting `.xpkg` is effectively **two OCI images in one**:
- An outer OCI image containing the YAML metadata
- An embedded inner OCI image containing the controller binary

### Summary

| Package type | Dockerfile? | Build tool |
|---|---|---|
| Configuration | **No** | `crossplane xpkg build` only — no Docker needed |
| Function / Provider | **Yes** | `docker build` first, then `crossplane xpkg build --embed-runtime-image` |

For your main job (writing Configurations) you **do not need Docker to build** — only to run `crossplane render` locally, which pulls function images to execute the pipeline dry-run.

---

## Pushing packages — any OCI tool works

Once a package is built, it is a **standard OCI image**. Pushing it does not require the Crossplane CLI — any OCI-compatible tool works:

### 1. Crossplane CLI (recommended — simplest)
```bash
crossplane xpkg push --package=my-platform.xpkg docker.io/yourorg/my-platform:v0.1.0
```

### 2. Docker — load tarball and push
```bash
docker load < my-platform.xpkg
docker tag <image-id> docker.io/yourorg/my-platform:v0.1.0
docker push docker.io/yourorg/my-platform:v0.1.0
```

### 3. crane — lightweight OCI tool (popular in CI)
```bash
crane push my-platform.xpkg docker.io/yourorg/my-platform:v0.1.0
```

### 4. skopeo — common in enterprise / air-gapped environments
```bash
skopeo copy oci-archive:my-platform.xpkg docker://docker.io/yourorg/my-platform:v0.1.0
```

The reason all these work: `crossplane xpkg build` produces a **standard OCI image**. The Crossplane-specific content is inside the layers — the image format itself is 100% standard OCI.

### Build vs push — which tool is required?

| Task | Required tool |
|---|---|
| **Build** a Configuration package | `crossplane xpkg build` — the only practical option |
| **Build** a Function/Provider package | `docker build` + `crossplane xpkg build --embed-runtime-image` |
| **Push** any package | Any OCI tool: `crossplane xpkg push`, `docker push`, `crane push`, `skopeo copy` |

The asymmetry:
- **Build** → Crossplane CLI is required (it understands the special OCI layer format and validates YAML)
- **Push** → any OCI tool works (it's a standard image once built)

In practice most teams use **Crossplane CLI for local dev** and **crane or skopeo in enterprise CI** where Docker daemon is not available (rootless containers, air-gapped environments).

---

## Authenticating with private registries

There are two separate login scenarios — pushing from your machine/CI, and pulling inside the cluster. They use different mechanisms.

### Pushing — login on your machine or CI

Standard Docker/OCI login, nothing Crossplane-specific. After logging in, `crossplane xpkg push`, `docker push`, and `crane push` all pick up credentials automatically from `~/.docker/config.json`.

```bash
# Docker Hub
docker login docker.io

# GitHub Container Registry
echo $GHCR_TOKEN | docker login ghcr.io -u youruser --password-stdin

# AWS ECR
aws ecr get-login-password --region eu-central-1 | \
  docker login --username AWS --password-stdin \
  123456789.dkr.ecr.eu-central-1.amazonaws.com
```

### Pulling — the cluster needs credentials

This is the Crossplane-specific part. The Crossplane package manager runs inside the cluster and needs credentials to pull from your private registry. You provide them via a Kubernetes image pull secret referenced in the package CR.

**Step 1 — create the pull secret in `crossplane-system`:**

```bash
kubectl create secret docker-registry my-registry-secret \
  --namespace crossplane-system \
  --docker-server=ghcr.io \
  --docker-username=youruser \
  --docker-password=$GHCR_TOKEN
```

**Step 2 — reference it in your Provider / Function / Configuration:**

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-platform
spec:
  package: ghcr.io/mycompany/my-platform:v0.1.0
  packagePullSecrets:
    - name: my-registry-secret   # must be in crossplane-system namespace
```

The `packagePullSecrets` field works the same way on all three package types: `Provider`, `Function`, and `Configuration`.

### Summary

| Scenario | How |
|---|---|
| Push from laptop / CI | `docker login` → credentials in `~/.docker/config.json` → used by any OCI push tool automatically |
| Pull in cluster (private registry) | Create `docker-registry` Secret in `crossplane-system` → reference in `spec.packagePullSecrets` |
| Pull in cluster (public registry) | No credentials needed |

---

## Practical tips

**Always tag with a specific version — never use `latest`**

```bash
# Bad — unpredictable, no revision history
spec:
  package: docker.io/yourorg/my-platform:latest

# Good — deterministic, reproducible
spec:
  package: docker.io/yourorg/my-platform:v0.1.0
```

**Use `--ignore` to keep the package small**

Only the files needed by Crossplane should be in the image. Exclude docs, examples, and test files:

```bash
crossplane xpkg build \
  --package-root=. \
  --output=my-platform.xpkg \
  --ignore="examples/**" \
  --ignore="tests/**" \
  --ignore="*.md" \
  --ignore=".github/**"
```

**Check what's inside a package**

Since `.xpkg` is an OCI tarball you can inspect it:

```bash
# Rename and extract
cp my-platform.xpkg my-platform.tar
mkdir inspect && tar -xf my-platform.tar -C inspect/
ls inspect/
```

---

## MRD (Managed Resource Definition)

When you install a Provider, it typically registers many CRDs — one per managed resource type (Bucket, Table, Cluster, etc.). In v2, Crossplane uses **Managed Resource Definitions (MRDs)** so you can **selectively activate** only the resource types you need.

### MR vs MRD — terminology

| Term | Meaning |
|---|---|
| **MR** (Managed Resource) | A concrete instance — e.g. one `Bucket` CR, one `Table` CR. The actual resource you create. |
| **MRD** (Managed Resource Definition) | The schema/definition for that resource type — the CRD. In v2, MRDs can be activated or left inactive. |

### Why MRDs matter

Large providers (e.g. provider-aws) can install 100+ CRDs. Each CRD consumes ~3 MiB of API server memory. If you only need S3 Buckets and DynamoDB Tables, MRDs let you activate just those — the other 98+ CRDs never get created.

### How to use MRDs

Create a **ManagedResourceActivationPolicy (MRAP)** to control which MRDs become active:

```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: ManagedResourceActivationPolicy
metadata:
  name: aws-s3-and-dynamodb-only
spec:
  activationPolicy: Manual
  resourceRefs:
    - name: buckets.s3.aws.m.upbound.io
    - name: tables.dynamodb.aws.m.upbound.io
```

Without an MRAP, the provider activates all MRDs by default. See [Managed Resources & Providers](./04_Managed_Resources_Providers.md#managed-resource-definitions-mrds--v2) for the full MRD section.

---

## Who creates what in practice

Understanding which package types you actually need to create vs just install is key to knowing where to focus your time.

**Providers — you install, you don't write**

Providers are written in Go and contain a full Kubernetes controller that reconciles managed resources against a real cloud API (AWS, GCP, Azure, ...). This is serious infrastructure engineering. You simply **install** existing providers from the Upbound Marketplace via a `Provider` CR. The only teams that write providers from scratch are:
- Upbound and crossplane-contrib maintainers (the official ones)
- Companies with a niche internal system that has no existing provider

**Functions — you install (occasionally write)**

Most teams use existing published functions: `function-patch-and-transform`, `function-kcl`, `function-go-templating`. Writing a custom function requires Go or another supported runtime and is relatively uncommon. When you use a Configuration's `dependsOn`, the function is auto-installed — you don't even have to think about it.

**Configurations — this is your main job**

As a platform engineer using Crossplane, writing Configurations is your primary task. You define XRDs (the API), Compositions (the implementation), wire them in `crossplane.yaml`, and publish the OCI image. Your app teams consume it by creating XR instances.

### Summary table

| Task | Who does it |
|---|---|
| Write a Provider | Upbound / open source maintainers |
| Install a Provider | You (platform engineer), or auto via `dependsOn` |
| Write a Function | Upbound / open source (mostly) — you rarely |
| Install a Function | Auto-installed via Configuration `dependsOn` |
| **Write a Configuration** | **You — this is your main job** |
| Install a Configuration | App teams / GitOps pipeline |
| Create XR instances | App teams / developers |

The analogy holds perfectly with Terraform: you don't write the AWS provider — HashiCorp does. You just declare `required_providers` and write modules. In Crossplane: you don't write providers — Upbound does. You declare `dependsOn` and write Configurations.

---

## See also

- [Packages & CLI](./07_Packages_CLI.md) — overview of packages and `crossplane render`
- [Configurations](./13_Configurations.md) — building Configuration packages in depth
- [Upbound Marketplace](./11_Upbound_Marketplace.md) — discovering published packages
- [Crossplane CLI reference](https://docs.crossplane.io/latest/cli/) — full CLI documentation
