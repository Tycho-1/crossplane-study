# Managed Resources & Providers

## Providers

**Providers** are Crossplane packages that:

- Install **CRDs** (Managed Resource Definitions) for an external system (e.g. AWS S3, RDS, Azure Storage).
- Run a **controller** that reconciles those CRs to the real APIs (create/update/delete).

Examples: [Provider AWS (Upjet)](https://github.com/crossplane-contrib/provider-upjet-aws), [Provider Azure](https://github.com/upbound/provider-azure), [Provider GCP](https://github.com/upbound/provider-gcp), [Provider Kubernetes](https://github.com/crossplane-contrib/provider-kubernetes).

### Install a provider

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws-s3
spec:
  package: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0
```

- Use a **fully qualified** package URL (required in v2).
- Optional: `packagePullPolicy`, `revisionActivationPolicy`, `revisionHistoryLimit`, `packagePullSecrets`, `runtimeConfigRef`.

### Verify

```bash
kubectl get providers
kubectl get providerrevisions
kubectl get crd | grep aws
```

---

## Provider configuration (credentials)

Each provider documents how to authenticate. For **Upbound AWS** providers, you use **ProviderConfig** or **ClusterProviderConfig** that point at a Kubernetes **Secret**.

### ClusterProviderConfig (cluster-wide)

Applies to MRs in any namespace unless they override with `providerConfigRef`:

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

### ProviderConfig (namespaced)

Applies to MRs in the **same namespace**:

```yaml
apiVersion: aws.m.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  namespace: default
  name: my-aws-config
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-secret
      key: creds
```

### Referencing in MRs

- Omit `providerConfigRef` → default is `name: default`, `kind: ClusterProviderConfig`.
- Explicit:

```yaml
spec:
  providerConfigRef:
    name: default
    kind: ClusterProviderConfig
  # or for namespaced:
  #   name: my-aws-config
  #   kind: ProviderConfig
  forProvider:
    region: eu-central-1
```

---

## Managed Resources (MRs)

**Managed resources** are the CRs that represent one external resource (e.g. one S3 bucket, one RDS instance). The provider controller reconciles them.

### v2 style: namespaced MRs

```yaml
apiVersion: s3.aws.m.upbound.io/v1beta1
kind: Bucket
metadata:
  namespace: default
  name: my-app-bucket
spec:
  forProvider:
    region: eu-central-1
  # optional:
  # providerConfigRef:
  #   name: default
  #   kind: ClusterProviderConfig
```

- **Status:** `SYNCED` / `READY` (and conditions) reflect reconciliation and readiness.
- **Deletion:** Removing the MR triggers deletion of the cloud resource (provider-dependent; some support deletion policies).

### Common MR patterns

- **forProvider:** provider-specific spec (region, tags, size, etc.).
- **providerConfigRef:** which ProviderConfig/ClusterProviderConfig to use.
- **writeConnectionSecretToRef:** (if supported) write connection info to a Secret (e.g. bucket name, endpoint).

---

## Managed Resource Definitions (MRDs) – v2

**Managed Resource Definitions (MRDs)** are lightweight abstractions over Kubernetes CRDs. In v2, providers ship MRDs instead of installing all CRDs immediately. You control which MRDs become **active** — only active MRDs create a real CRD in the cluster.

### Why MRDs matter

Large providers (e.g. provider-aws) can install 100+ CRDs. Each CRD consumes ~3 MiB of API server memory. If you only need S3 Buckets and DynamoDB Tables, you don't want 98 other CRDs. MRDs let you **selectively activate** only the resources you use.

### How it works

- **MRD state:** `Active` → CRD exists in cluster; `Inactive` → no CRD, resource type is unavailable
- **ManagedResourceActivationPolicy (MRAP):** controls which MRDs become active via pattern matching

### Example — activate only S3 and DynamoDB

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

Without an MRAP, the provider activates all MRDs by default (legacy behaviour). With an MRAP, only matching MRDs become active.

### MR vs MRD — terminology

| Term | Meaning |
|---|---|
| **MR** (Managed Resource) | A concrete instance — e.g. one `Bucket` CR, one `Table` CR |
| **MRD** (Managed Resource Definition) | The schema/definition for that resource type — the CRD. In v2, MRDs can be selectively activated. |

See [Managed Resource Definitions](https://docs.crossplane.io/latest/managed-resources/managed-resource-definitions/) and [Disabling Unused Managed Resources](https://docs.crossplane.io/latest/guides/disabling-unused-managed-resources/) in the official docs.

---

## Runtime configuration (DeploymentRuntimeConfig)

**DeploymentRuntimeConfig** (beta) customizes the **Deployment** of Provider (and Function) pods: args, resources, tolerations, etc.

Example – pass an extra flag to the provider controller:

```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: DeploymentRuntimeConfig
metadata:
  name: my-provider-runtime
spec:
  deploymentTemplate:
    spec:
      template:
        spec:
          containers:
            - name: package-runtime
              args:
                - --some-flag
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-aws-s3
spec:
  package: xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0
  runtimeConfigRef:
    name: my-provider-runtime
```

---

## RBAC for composing non-Crossplane resources (v2)

If your **Composition** creates resources that are **not** Crossplane MRs or XRs (e.g. `postgresql.cnpg.io` Cluster, or core Deployment/Service), Crossplane’s service account must have RBAC for those types.

Use **ClusterRole aggregation**:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cnpg:aggregate-to-crossplane
  labels:
    rbac.crossplane.io/aggregate-to-crossplane: "true"
rules:
  - apiGroups: ["postgresql.cnpg.io"]
    resources: ["clusters"]
    verbs: ["*"]
```

The label `rbac.crossplane.io/aggregate-to-crossplane: "true"` makes this role aggregate into Crossplane’s main ClusterRole.

---

## Upgrade, remove, troubleshoot

- **Upgrade provider:** change `spec.package` to a new version and apply; Crossplane creates a new ProviderRevision.
- **Remove provider:** delete the Provider. Remove or orphan MRs first to avoid leaving cloud resources without a controller; see [Troubleshooting](https://docs.crossplane.io/latest/guides/troubleshoot-crossplane/).
- **Unhealthy provider:** `kubectl describe providerrevisions` and check events; often credentials or dependency/version issues.
- **Two AWS family providers:** Do not install both **crossplane-contrib** and **Upbound** `provider-family-aws` on one cluster — see [AWS family: contrib vs Upbound](./18_AWS_Family_Contrib_vs_Upbound.md).

---

## Next

- [Composition (XRDs, Compositions, Functions)](./05_Composition_XRDs_Functions.md) – Custom APIs and function pipelines.
- [Packages & CLI](./07_Packages_CLI.md) – Functions, Configurations, `crossplane` CLI.
- Practice: [examples/01-simple-s3-bucket](../examples/01-simple-s3-bucket/), [examples/02-provider-config-aws](../examples/02-provider-config-aws/).
