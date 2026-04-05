# Installation & Setup

## Prerequisites

- **Kubernetes cluster** (supported version, see [Kubernetes releases](https://kubernetes.io/releases/patch-releases/#support-period))
- **Helm** v3.2.0 or later
- (Optional) Local cluster: [Kind](https://kind.sigs.k8s.io/)

---

## Install Crossplane

Crossplane runs as a pod in your cluster and enables **Provider**, **Function**, and **Configuration** packages.

### 1. Add Helm repo and install

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm install crossplane \
  --namespace crossplane-system \
  --create-namespace \
  crossplane-stable/crossplane
```

### 2. Verify

```bash
kubectl get pods -n crossplane-system
```

You should see `crossplane` and `crossplane-rbac-manager` (if enabled) running.

---

## Installation Options

### Customize the Helm chart

Use `--set` or a `values.yaml`. See the [chart README](https://github.com/crossplane/crossplane/blob/main/cluster/charts/crossplane/README.md#configuration).

### Feature flags

Some features are behind flags. Set them via Helm `args` (e.g. in `values.yaml` or `--set args='{...}'`).

| Status | Flag | Description |
|--------|------|-------------|
| Beta | `--enable-deployment-runtime-configs` | DeploymentRuntimeConfigs for Provider/Function runtime |
| Beta | `--enable-usages` | Usages |
| Beta | `--enable-realtime-compositions` | Real-time compositions |
| Alpha | `--enable-dependency-version-upgrades` | Auto version upgrades for package dependencies |
| Alpha | `--enable-function-response-cache` | Cache composition function responses |
| Alpha | `--enable-signature-verification` | Package signature verification (ImageConfig) |

**Example – enable beta features:**

```yaml
# values.yaml
args:
  - --enable-deployment-runtime-configs
  - --enable-usages
```

```bash
helm install crossplane crossplane-stable/crossplane \
  -n crossplane-system --create-namespace \
  -f values.yaml
```

---

## Install a Provider

Providers add **managed resource** CRDs and a controller. Example: **provider-aws-s3** (S3 only).

### 1. Install the provider

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws-s3
spec:
  package: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0
EOF
```

### 2. Wait until healthy

```bash
kubectl get providers
# NAME                INSTALLED   HEALTHY   PACKAGE                                              AGE
# provider-aws-s3     True         True       xpkg.crossplane.io/.../provider-aws-s3:v2.0.0        1m
```

### 3. (Optional) Install with Helm

You can add providers at Crossplane install time:

```bash
helm install crossplane crossplane-stable/crossplane \
  -n crossplane-system --create-namespace \
  --set provider.packages='{xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0}'
```

---

## Configure Provider Credentials (AWS example)

The AWS provider needs credentials. Use a **ProviderConfig** (or **ClusterProviderConfig**) that references a **Secret**.

### 1. Create secret (e.g. from AWS CLI profile)

```bash
# Example: use existing AWS profile
kubectl create secret generic aws-secret \
  -n crossplane-system \
  --from-file=creds=$HOME/.aws/credentials
```

Or create a secret with access key and secret key (see your provider docs).

### 2. Create ProviderConfig

**ClusterProviderConfig** (cluster-wide, default for MRs that don’t specify a config):

```yaml
apiVersion: aws.m.upbound.io/v1beta1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-secret
      key: creds
```

**ProviderConfig** (namespaced): same `spec`, but `metadata.namespace` set. MRs in that namespace can reference it with `providerConfigRef`.

### 3. Use in a managed resource

If you use the default ClusterProviderConfig named `default`, MRs don’t need to set `providerConfigRef`. Otherwise:

```yaml
spec:
  providerConfigRef:
    name: default
    kind: ClusterProviderConfig
  # or for namespaced: kind: ProviderConfig
  forProvider:
    region: eu-central-1
```

---

## Install a Composition Function

Compositions use **functions** (e.g. Patch & Transform) in a pipeline. Install the function as a **Function** package:

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.crossplane.io/crossplane-contrib/function-patch-and-transform:v0.8.2
EOF
```

```bash
kubectl get functions
# NAME                            INSTALLED   HEALTHY   PACKAGE     AGE
# function-patch-and-transform    True        True      xpkg...     1m
```

---

## Pre-release (master) install

For development only:

```bash
helm repo add crossplane-master https://charts.crossplane.io/master/
helm repo update
helm install crossplane -n crossplane-system --create-namespace crossplane-master/crossplane --devel
```

Do not use the master channel in production.

---

## Next Steps

- [Managed Resources & Providers](./04_Managed_Resources_Providers.md) – Create MRs, ProviderConfig, RBAC.
- [Composition](./05_Composition_XRDs_Functions.md) – XRDs, Compositions, and function pipelines.
- Try [examples/01-simple-s3-bucket](../examples/01-simple-s3-bucket/) after provider and credentials are set.
