# VPC + EKS with Kustomize: Parameterized Example

This document explains how to deploy **both** a VPC (Network) and an EKS cluster using the Upbound Configurations, with **Kustomize** ensuring the shared `id` parameter stays in sync. No more manual copy-paste or mismatched IDs.

**Full prerequisite order** (family provider, **both** provider configs, Configurations, matching `id`): see **[README → Install and use → “Before everything works”](./README.md#install-and-use)**. Easy to miss: **namespaced** [provider-config-default.yaml](./provider-config-default.yaml) for Network, not only `ClusterProviderConfig`. **AWS console subnet/VPC “names”** are tags — see [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md#aws-console-names-for-vpcs-and-subnets).

---

## The Problem: Why `id` Must Match

The EKS configuration does **not** create a VPC. It uses **label selectors** to find subnets:

```
EKS composition looks for subnets with label:
  networks.aws.platform.upbound.io/network-id = <your id>
```

The Network (VPC) configuration **labels** its subnets with that same `id`:

```
Network composition labels subnets with:
  networks.aws.platform.upbound.io/network-id = <your id>
```

So **both** resources must use the **same** `id`. If they differ, EKS won't find any subnets and the cluster will never become Ready.

---

## The Solution: Single Source of Truth

We use **Kustomize replacements** to inject the same `environmentId` into both the Network and EKS XRs from a ConfigMap. Change one value, both get updated.

```
env-config (ConfigMap)
    │
    ├── environmentId: "prod-cluster-01"
    │       │
    │       ├──► Network.spec.parameters.id
    │       └──► EKS.spec.parameters.id
    │
    └── region: "eu-central-1"
            │
            ├──► Network.spec.parameters.region
            └──► EKS.spec.parameters.region
```

---

## File Layout

```
configuration-aws-eks/
├── configuration-aws-network-install.yaml   # Installs Network Configuration
├── configuration-aws-eks-install.yaml       # Installs EKS Configuration
├── kustomization.yaml                        # Kustomize overlay + replacements
├── kustomize/
│   ├── env-config.yaml                      # ← Edit this: environmentId, region
│   ├── network-xr.yaml                      # Network (VPC) XR template
│   └── eks-xr.yaml                          # EKS XR template
├── network-xr-example.yaml                  # Standalone Network (no Kustomize)
├── eks-xr-example.yaml                      # Standalone EKS (requires Network elsewhere)
└── VPC-AND-EKS-WITH-KUSTOMIZE.md            # This file
```

---

## How It Works

### 1. env-config.yaml — Single Source of Truth

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: env-config
  namespace: default
data:
  environmentId: "prod-cluster-01"   # ← Change this for dev/staging/prod
  region: "eu-central-1"             # ← Change this for your region
```

### 2. Kustomize Replacements

The `kustomization.yaml` defines replacements that copy values from the ConfigMap into both XRs:

```yaml
replacements:
  - source:
      kind: ConfigMap
      name: env-config
      fieldPath: data.environmentId
    targets:
      - select:
          kind: Network
        fieldPaths:
          - spec.parameters.id
      - select:
          kind: EKS
        fieldPaths:
          - spec.parameters.id
  # ... same for region
```

### 3. Apply Order

Kustomize outputs all resources in one manifest. Apply order matters for Crossplane:

1. **provider-family-aws** — required for Network (`ec2.aws.m.upbound.io`). Wait until HEALTHY. See [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md) if you get "no matches for kind" errors.
2. **Secret** `aws-secret` (key `creds`) in `crossplane-system`.
3. **ClusterProviderConfig** — [cluster-provider-config.yaml](./cluster-provider-config.yaml) **after** family CRDs exist (`kubectl get crd clusterproviderconfigs.aws.m.upbound.io`).
4. **ProviderConfig** `default` in the XR namespace — [provider-config-default.yaml](./provider-config-default.yaml). Required for Network composed MRs; without it: `ProviderConfig "default" not found`.
5. **Configurations** (network, eks) — install, wait HEALTHY.
6. **ConfigMap** — env-config (if using kustomize).
7. **Network** XR — VPC + subnets, labels `network-id=<id>`.
8. **EKS** XR — same `<id>` as Network (Kustomize keeps them aligned).

Kustomize does not guarantee order. To be safe, apply in two steps:

```bash
# Step 1: Install Configurations, wait for HEALTHY
kubectl apply -f configuration-aws-network-install.yaml
kubectl apply -f configuration-aws-eks-install.yaml
kubectl get configuration   # wait until both HEALTHY

# Step 2: Apply Network, ConfigMap, EKS
kubectl apply -k .
# Or: kubectl kustomize . | kubectl apply -f -
```

Or apply everything at once — Crossplane will retry until the Network is Ready and EKS can find the subnets.

---

## Usage

### Deploy

```bash
cd configuration-aws-eks
kubectl apply -k .
```

### Customize for Another Environment

Edit `kustomize/env-config.yaml`:

```yaml
data:
  environmentId: "dev-cluster-01"
  region: "us-east-1"
```

Then re-apply:

```bash
kubectl apply -k .
```

### Verify

```bash
# Configurations
kubectl get configuration

# Network (VPC) — wait until Ready
kubectl get network.aws.platform.upbound.io

# EKS — wait until Ready
kubectl get eks.aws.platform.upbound.io

# Composed resources
kubectl get vpc.ec2.aws.upbound.io
kubectl get cluster.eks.aws.upbound.io
```

---

## Kustomize Overlays for Multiple Environments

For dev/staging/prod, create overlays that patch `env-config`:

```
configuration-aws-eks/
├── kustomization.yaml
├── kustomize/
│   ├── env-config.yaml
│   ├── network-xr.yaml
│   └── eks-xr.yaml
└── overlays/
    └── dev/
        ├── kustomization.yaml
        └── env-config-patch.yaml
```

Example `overlays/dev/kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../.
patches:
  - path: env-config-patch.yaml
    target:
      kind: ConfigMap
      name: env-config
```

Example `overlays/dev/env-config-patch.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: env-config
  namespace: default
data:
  environmentId: "dev-cluster-01"
  region: "us-west-2"
```

Then: `kubectl apply -k overlays/dev/`

---

## Summary

| What | How |
|------|-----|
| **Link VPC ↔ EKS** | Same `id` in both — Network labels subnets, EKS selects by label |
| **Parameterize** | ConfigMap `env-config` holds `environmentId` and `region` |
| **Keep in sync** | Kustomize replacements inject ConfigMap values into both XRs |
| **Change env** | Edit `kustomize/env-config.yaml`, re-apply |

---

## See also

- [XR-PARAMETERS.md](./XR-PARAMETERS.md) — Full EKS parameter reference
- [HOW-IT-WORKS.md](./HOW-IT-WORKS.md) — How one EKS XR becomes many AWS resources
- [configuration-aws-network](https://github.com/upbound/configuration-aws-network) — Network Configuration source
