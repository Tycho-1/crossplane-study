# Open Forms — Crossplane Configuration (user guide)

Deploy **[Open Forms](https://open-forms.readthedocs.io/)** on Kubernetes using a **single Crossplane composite resource** (`OpenForms`). This Configuration composes plain Kubernetes objects (Deployments, Services, ConfigMaps, PVC, etc.); it does **not** install a Helm `Release` at runtime.

| | |
|--|--|
| **Upstream chart** | [maykinmedia/openforms](https://github.com/maykinmedia/charts/tree/main/charts/openforms) |
| **Chart version** | 1.12.0 |
| **App / image** | 3.3.9 (`openformulieren/open-forms`) |
| **API** | `openforms.example.org/v1alpha1` · kind **`OpenForms`** |

For **maintainers** (regenerating the Composition, upgrading the chart port, building the package), see **[README.md](./README.md)**.

---

## Introduction

### What you provide

- A **PostgreSQL** database reachable from the cluster.
- **Redis** (or compatible) for Celery **broker** and **result backend** — this Configuration uses the **external Redis** layout (no Bitnami Redis subchart in the composed manifests).
- **Secrets** and URLs via `spec.parameters` on the **OpenForms** resource.

### What Crossplane creates

ServiceAccount, Secret, ConfigMaps (app + nginx), PVC, Services, Deployments (web, nginx, worker, beat, flower) — matching the embedded chart render for **external Redis**. See **[README.md](./README.md)** for the full scope and limitations.

---

## Redis and Helm subcharts — who creates what?

The [upstream chart](https://github.com/maykinmedia/charts/blob/main/charts/openforms/README.md) declares **dependencies** on Bitnami **redis** and **common** (see chart `Chart.yaml`). When you run **`helm install`**, Helm renders and applies **Open Forms** plus the **Redis** subchart (StatefulSet, Services, ConfigMaps for Redis, etc.) and uses **common** as a library chart (helpers), not as extra standalone app resources.

**This Crossplane Configuration does *not* create Redis (or any subchart objects) today.**

| Approach | What creates Redis | What this Configuration composes |
|----------|---------------------|----------------------------------|
| **Current design (external Redis)** | **You** — e.g. a separate Helm release, Redis Operator, managed cloud Redis, or a small manifest you apply by hand. | Only the **Open Forms** chart resources that remain when Redis is turned off in the chart (`tags.redis=false` in `helm template`). Celery URLs (`settingsCeleryBrokerUrl` / `settingsCeleryResultBackend`) must point at **that** Redis. |
| **“Everything in one Composition”** | **Crossplane** would compose Redis too — by **embedding** the same YAML Helm would render for the Redis subchart (from `helm template` **with** the subchart enabled). | You would add those objects to **`apis/composition.yaml`** (or switch the generator to [reference/helm-template-openforms-1.12.0-full.yaml](./reference/helm-template-openforms-1.12.0-full.yaml)) and add XRD parameters for Redis storage, password, etc. |

**Why external Redis was chosen for the first port:** fewer composed objects, no Bitnami Redis upgrade drift inside this repo, and many teams already run a shared Redis. **Trade-off:** you must run Redis yourself and keep URLs consistent.

**If you want Redis inside the same XR:** extend the Composition with the StatefulSet/Services/… from the **full** reference render, regenerate, and expose the few Redis-related settings you need on `spec.parameters`. That is the same maintenance model as the rest of the chart port (diff on chart bump). The **`common`** subchart does not add separate Deployments; it is template helpers used by the Redis chart.

---

## Requirements

### Cluster

| Requirement | Notes |
|-------------|--------|
| **Crossplane** | v2+ with **Composition pipeline** mode |
| **Functions** | [`function-patch-and-transform`](https://github.com/crossplane-contrib/function-patch-and-transform), [`function-auto-ready`](https://github.com/crossplane-contrib/function-auto-ready) — declared in [crossplane.yaml](./crossplane.yaml) `dependsOn` |
| **RBAC** | Crossplane must be allowed to create the composed API kinds in the target namespace (see [examples/05-app-with-deployment](../examples/05-app-with-deployment/README.md)) |
| **Storage** | RWX PVC for app media; default size is `1Gi` (configurable) |

### Outside the platform

| Dependency | Description |
|------------|----------------|
| **PostgreSQL** | `DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` via parameters |
| **Redis** | Broker and result backend URLs (`redis://…`) |

---

## Quickstart

### 1. Install the Configuration package

Install the **Configuration** from your OCI registry (after you build and push the package), or apply the manifests directly:

```bash
# Option A — from registry (edit image in examples/install-from-registry.yaml first)
kubectl apply -f examples/install-from-registry.yaml

# Option B — manifests only (no package)
kubectl apply -f apis/definition.yaml
kubectl apply -f apis/composition.yaml
```

Wait until the **Configuration** is **HEALTHY** (if using a package) and **Functions** from `dependsOn` are installed.

### 2. Install functions (if not pulled by the package)

Use the same pattern as [examples/05-app-with-deployment/functions.yaml](../examples/05-app-with-deployment/functions.yaml) — versions should satisfy [crossplane.yaml](./crossplane.yaml).

### 3. Create an OpenForms instance

**Important:** `metadata.name` **must** be **`demo`** (see [Resource name](#resource-name-metadata)).

```yaml
apiVersion: openforms.example.org/v1alpha1
kind: OpenForms
metadata:
  name: demo
  namespace: default
spec:
  parameters:
    namespace: openforms
    settingsSecretKey: "<generate-a-long-random-secret>"
    settingsBaseUrl: "https://openforms.example.com"
    settingsDatabaseHost: "postgres.example.svc.cluster.local"
    settingsDatabaseName: "openforms"
    settingsDatabaseUsername: "openforms"
    settingsDatabasePassword: "<secret>"
    settingsCeleryBrokerUrl: "redis://redis:6379/0"
    settingsCeleryResultBackend: "redis://redis:6379/1"
    settingsNumProxies: "1"
    replicaCount: 2
    workerReplicaCount: 2
    persistenceSize: 1Gi
    persistenceStorageClassName: ""
```

Copy from **[openforms-xr-example.yaml](./openforms-xr-example.yaml)** and adjust values.

```bash
kubectl apply -f openforms-xr-example.yaml
kubectl get openforms -n default
kubectl get pods -n openforms
```

---

## Resource name (`metadata`)

| Field | Value | Why |
|-------|--------|-----|
| **`metadata.name`** | **`demo`** (required) | The Composition embeds manifests rendered with Helm release name `demo` (`demo-openforms`, `ALLOWED_HOSTS`, nginx `proxy_pass`, etc.). |

To use a **different** name, the Configuration must be **regenerated** from a new `helm template` reference (see [README.md](./README.md)).

---

## Parameters (`spec.parameters`)

All user-facing settings are under **`spec.parameters`**.

### Coverage vs Helm `values.yaml`

**Yes — Helm exposes strictly more than this Configuration.** The chart’s [`values.yaml`](https://github.com/maykinmedia/charts/blob/main/charts/openforms/values.yaml) is the **full** upstream surface (dozens of keys: ingress, images, probes, resources, Redis subchart, `flower`, Sentry, autoscaling, etc.). This port only promotes **some** of those into Crossplane parameters; the rest stay at whatever the **reference** `helm template` used when [apis/composition.yaml](./apis/composition.yaml) was generated (defaults baked into embedded YAML). To expose another chart value, add a field to the XRD and a Patch & Transform rule (or extend [scripts/generate_composition.py](./scripts/generate_composition.py)).

### Values table

| Key | Type | Default | Required | Description |
|-----|------|---------|----------|-------------|
| `namespace` | string | `openforms` | no | Namespace for **all** composed resources. Embedded strings (`ALLOWED_HOSTS`, nginx upstreams) assume **`openforms`**; changing it requires a **regenerated** Composition (see maintainer README). |
| `settingsSecretKey` | string | — | **yes** | Django `SECRET_KEY` (also `stringData.SECRET_KEY`). Generate a strong value (e.g. [djecrety](https://djecrety.ir/) as in the [upstream chart README](https://github.com/maykinmedia/charts/blob/main/charts/openforms/README.md)). |
| `settingsBaseUrl` | string | — | **yes** | Public base URL (`BASE_URL`), e.g. `https://forms.example.com`. |
| `settingsDatabaseHost` | string | — | **yes** | PostgreSQL host (`DB_HOST`). |
| `settingsDatabaseName` | string | — | **yes** | Database name (`DB_NAME`). |
| `settingsDatabaseUsername` | string | — | **yes** | Database user (`DB_USER`). |
| `settingsDatabasePassword` | string | — | **yes** | Database password (`DB_PASSWORD` in Secret). Prefer secrets management in production. |
| `settingsCeleryBrokerUrl` | string | — | **yes** | Celery broker URL (`CELERY_BROKER_URL`), e.g. `redis://redis:6379/0`. |
| `settingsCeleryResultBackend` | string | — | **yes** | Celery result backend (`CELERY_RESULT_BACKEND`), e.g. `redis://redis:6379/1`. |
| `settingsNumProxies` | string | `"1"` | no | `NUM_PROXIES` in ConfigMap. Use **`"2"`** when exposing the app behind an Ingress (see [upstream chart](https://github.com/maykinmedia/charts/blob/main/charts/openforms/README.md)). |
| `replicaCount` | integer | `2` | no | Replicas for the main web/uWSGI Deployment (`demo-openforms`). |
| `workerReplicaCount` | integer | `2` | no | Replicas for the Celery worker Deployment. |
| `persistenceSize` | string | `1Gi` | no | PVC size for shared media volume. |
| `persistenceStorageClassName` | string | `""` | no | Optional StorageClass for the PVC; empty uses cluster default. |

### Mapping to Helm chart values

If you know the [upstream `values.yaml`](https://github.com/maykinmedia/charts/blob/main/charts/openforms/values.yaml), this is a rough mapping:

| Helm value | OpenForms parameter |
|------------|---------------------|
| `settings.secretKey` | `settingsSecretKey` |
| `settings.baseUrl` | `settingsBaseUrl` |
| `settings.database.host` / `name` / `username` / `password` | `settingsDatabaseHost` / `…Name` / `…Username` / `…Password` |
| `settings.celery.brokerUrl` | `settingsCeleryBrokerUrl` |
| `settings.celery.resultBackend` | `settingsCeleryResultBackend` |
| `settings.numProxies` | `settingsNumProxies` (string) |
| `replicaCount` | `replicaCount` |
| `worker.replicaCount` | `workerReplicaCount` |
| `persistence.size` | `persistenceSize` |
| `persistence.storageClassName` | `persistenceStorageClassName` |

**Not** exposed as parameters in this Configuration (would require extending the XRD and Composition): Ingress, image tags, `flower`, Sentry, automatic `django-setup-configuration` job, embedded Bitnami Redis, and most other `values.yaml` keys. See [README.md](./README.md) to extend the port.

### Design: curated parameters vs full Helm values

You *can* mirror the entire chart `values.yaml` in the XRD, but it is usually **not** the best fit for how Crossplane is typically used.

| Idea | Notes |
|------|--------|
| **Curated API (usual Crossplane pattern)** | The composite resource exposes a **small, stable** set of fields; the Composition encodes defaults, wiring, and guardrails. Callers get a **platform** API, not a dump of every chart knob. |
| **Exposing “everything”** | Duplicates Helm’s surface in another schema, inflates the Composition (patches or functions), and shifts maintenance to your port **without** simplifying operations. The XR can end up as “Helm with extra steps.” |
| **Full Helm parity with minimal porting** | Often **[`provider-helm` `Release`](https://github.com/crossplane-contrib/provider-helm)** is the clearer Crossplane pattern: one managed resource, full `values`, upstream chart stays the source of truth. |
| **Hand-ported manifests (this repo)** | A **partial** parameter set + **embedded** defaults from a known `helm template` run is a common trade-off: simpler API, explicit drift when the chart changes. |

**Practical patterns:** (1) **Curated** parameters + baked-in reference render (what this Configuration does). (2) **Hybrid** — curated fields plus a small “advanced” escape hatch, not literally every key. (3) **Generated** XRD/Composition if you truly need wide coverage. (4) **`Release`** when tenants need the **full** `values.yaml` surface under Crossplane reconciliation.

**Bottom line:** A **narrow** `OpenForms` API matches Crossplane’s usual role as a **control-plane abstraction**; exposing **all** chart settings is valid for special cases (temporary parity, internal tooling) but fights that model — and **Helm `Release`** is often the better fit when “every knob” matters.

**Same idea for Terraform modules vs packaged Configurations (EKS, etc.):** see **[Terraform modules vs Crossplane Configurations](../guides/20_Terraform_Modules_vs_Crossplane_Configurations.md)**.

---

## Security

- `settingsSecretKey` and `settingsDatabasePassword` are **sensitive**. For production, prefer **external secrets** or **Vault** and wire values into the XR (or fork the Composition to use `Secret` references).
- Do **not** commit real secrets to Git.

---

## Django / Open Forms documentation

- **Open Forms docs:** [open-forms.readthedocs.io](https://open-forms.readthedocs.io/)
- **Upstream Helm chart README:** [maykinmedia/charts — openforms README](https://github.com/maykinmedia/charts/blob/main/charts/openforms/README.md)

---

## Publishing on Upbound Marketplace (or similar)

1. Build and push the package ([README.md — Build](./README.md#build-and-install-as-a-configuration-package)).
2. Point users to **this file (`USAGE.md`)** as the **parameter reference** and to **[README.md](./README.md)** for operational constraints.
3. In the listing, set **dependsOn** expectations: Crossplane v2, **function-patch-and-transform**, **function-auto-ready** (versions in [crossplane.yaml](./crossplane.yaml)).

---

## Support

- **Behaviour of the app** (Django, admin, forms): [Open Forms documentation](https://open-forms.readthedocs.io/).
- **This Configuration** (parameters, `demo` name, namespace): [README.md](./README.md) and this file.
