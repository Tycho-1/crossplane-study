# Crossplane for Kubernetes Applications (v2)

Crossplane is not only for **cloud infrastructure** (S3, RDS, EKS). In **v2**, you can use it to create and manage **normal Kubernetes applications** in the same cluster: Deployments, Services, ConfigMaps, Secrets, and any other Kubernetes or custom resources. This was a major v2 change: **Compositions can compose any Kubernetes resource**, not just Crossplane managed resources (MRs).

---

## Why use Crossplane for K8s applications?

- **Custom app APIs:** Define platform APIs like `XApp` or `XMicroservice`; developers create one XR and get a Deployment + Service + ConfigMap (or more) in one go.
- **Same workflow as infra:** Same tool (Crossplane), same patterns (XRDs, Compositions, functions), same GitOps and RBAC.
- **Full-stack abstractions:** One XR can create both **cloud resources** (e.g. RDS, S3) and **in-cluster resources** (Deployment, Service) together.
- **Policy and governance:** Enforce standards (image, replicas, resources) in the Composition; users only see the abstract API.

So Crossplane becomes a **control plane for both infrastructure and applications** in the cluster.

---

## v2: Compose any Kubernetes resource

In v1, Compositions could only create **Crossplane MRs** (and XRs). In **v2**:

- Compositions can create **any** Kubernetes resource:
  - **Built-in:** Deployment, Service, ConfigMap, Secret, Ingress, etc.
  - **Crossplane MRs:** Bucket, DBInstance, Cluster, etc.
  - **Third-party CRs:** e.g. CloudNativePG Cluster, Argo Application, Helm Release.
  - **Helm charts as managed releases:** see **[19 – Configuration from a Helm chart](./19_Configuration_From_Helm_Chart.md)** (`provider-helm` `Release` MR composed from your XR).

So you can define an **“App”** XR that composes only K8s resources (no cloud MRs), or mix cloud and in-cluster in one Composition.

---

## Granting Crossplane access to K8s resources

Crossplane already has access to:

- Resources from **Providers** (MRs).
- Resources from **XRDs** (XRs).

To compose **other** resource types (e.g. `apps/v1` Deployment, `v1` Service, or `postgresql.cnpg.io` Cluster), you must **grant Crossplane RBAC**. Use a **ClusterRole** that **aggregates** to Crossplane’s ClusterRole via the label:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: core-apps:aggregate-to-crossplane
  labels:
    rbac.crossplane.io/aggregate-to-crossplane: "true"
rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets"]
    verbs: ["*"]
  - apiGroups: [""]
    resources: ["services", "configmaps", "secrets", "pods"]
    verbs: ["*"]
```

Add rules for any API group/resource you want Crossplane to create or update. See [Composition – Grant access to composed resources](https://docs.crossplane.io/latest/composition/compositions/#grant-access-to-composed-resources).

---

## Example: XApp → Deployment + Service

A minimal “app” API that creates only Kubernetes resources (no cloud):

1. **XRD** – Define `XApp` with fields like `image`, `replicas`, `port`.
2. **Composition** – Use **function-patch-and-transform** (or CUE/KCL/Python) to output a **Deployment** and a **Service** from the XR spec.
3. **XR** – User creates one `XApp`; Crossplane creates the Deployment and Service in the same namespace.

**User creates:**

```yaml
apiVersion: example.org/v1alpha1
kind: XApp
metadata:
  namespace: default
  name: my-app
spec:
  image: nginx
  replicas: 2
  port: 80
```

**Crossplane creates** (via the Composition): a Deployment and a Service, both in `default`, owned by the XR.

This is what the **application examples** do: [05-app-with-deployment](../examples/05-app-with-deployment/) and [08-application](../examples/08-application/) – XApp that composes a Deployment and a Service using Patch & Transform, plus an optional RBAC ClusterRole for `apps` and `""` resources.

---

## Composition for applications vs Helm

Crossplane Composition (for apps) and **Helm** both turn “one logical thing” (an app) into **multiple Kubernetes resources** (Deployment, Service, etc.) and both support **parameterization**. The model is different.

### Similarities

| Aspect | Helm | Crossplane Composition (apps) |
|--------|------|-------------------------------|
| **Input** | Chart + values (e.g. `values.yaml`, `--set`) | XR spec (e.g. `image`, `replicas`, `port`) |
| **Output** | Rendered manifests (Deployment, Service, …) | Composed resources (Deployment, Service, …) |
| **Abstraction** | One “release” → many resources | One XR → many resources |
| **Parameterization** | `{{ .Values.image }}`, `{{ .Values.replicas }}` | Patches (e.g. FromCompositeFieldPath) or function logic |
| **GitOps** | Argo CD / Flux can sync Helm releases | Same tools can sync XRs; Crossplane reconciles continuously |

So both give you **one parameterized definition** that expands into a set of K8s resources.

### Differences

| Aspect | Helm | Crossplane Composition (apps) |
|--------|------|-------------------------------|
| **Model** | **Template + values** → render at install/upgrade time | **Custom API (XR)** → controller **reconciles** and applies composed resources |
| **Lifecycle** | `helm install` / `helm upgrade`; no controller watching the “release” as a CR in the API | XR is a **Kubernetes resource**; Crossplane **continuously** reconciles it and keeps composed resources in sync |
| **API** | No custom API; you run `helm install` with values | **Custom API** (e.g. `XApp`); users do `kubectl apply -f xapp.yaml` |
| **Templating** | Go templates in chart YAML | Composition **functions** (Patch & Transform, CUE, KCL, Python) – no Go templates in the chart sense |
| **Packaging** | Chart = versioned package (tgz, OCI); versioning and rollback built in | XRD + Composition + (optional) Functions; versioning via CompositionRevisions, no separate “chart package” |
| **Scope** | Apps only (K8s resources) | **Apps and/or cloud** – one Composition can output both K8s resources and Crossplane MRs (e.g. RDS + Deployment + Service) |
| **RBAC** | Who can run `helm install` / manage releases | Who can create/update **XRs** (e.g. `XApp`); fine-grained per custom API |

So: Helm is a **packaged template + values** applied at install/upgrade; Crossplane is a **custom resource type** with a **controller** that continuously reconciles and applies composed resources (and can mix in cloud MRs).

### When to use which

- **Helm** – Good when you want **versioned, portable charts** (e.g. from a registry), **Helm-native** tooling and rollback, and **template + values** is enough. Fits “install/upgrade this app” workflows.
- **Crossplane Composition (apps)** – Good when you want a **custom platform API** (e.g. `XApp`, `XMicroservice`) that **lives in the cluster** and is **reconciled continuously**, when you want **one abstraction for both app and infra** (e.g. XBackend → RDS + Deployment + Service), or when you already use Crossplane for infra and want the same model for apps.

You can also use **both**: e.g. Crossplane composes an XApp that creates a **Helm Release** (via provider-helm) so the “app” API is Crossplane, and the actual app manifests come from a Helm chart.

---

## Mixing cloud and in-cluster resources

You can compose **both** cloud MRs and normal K8s resources in one XR:

- **XBackend** → RDS Instance (MR) + Deployment + Service + Secret (connection string).
- **XDataApp** → S3 Bucket (MR) + Deployment (app that uses the bucket) + Service.

Define one XRD and one Composition whose function pipeline outputs both MRs and core/CR resources. Grant Crossplane RBAC for the non-MR resources (Deployment, Service, etc.).

---

## Relation to other concepts

| Concept | Cloud-only (v1 style) | Applications + cloud (v2) |
|--------|------------------------|----------------------------|
| Composed resources | MRs (and XRs) only | MRs + any K8s resource (Deployment, Service, CRs) |
| RBAC | Providers/XRDs only | Plus aggregate ClusterRole for Deployment, Service, etc. |
| Use case | Infra provisioning | App deployment, full-stack, custom platform APIs |

---

## Summary

- **v2:** Compositions can compose **any** Kubernetes resource, not only Crossplane MRs.
- Use Crossplane for **normal applications** in the cluster (Deployments, Services, ConfigMaps, etc.) by defining XRDs and Compositions that output those resources.
- **RBAC:** Add an aggregate ClusterRole so Crossplane can create/update the K8s resource types you use.
- **Examples:** [05-app-with-deployment](../examples/05-app-with-deployment/), [08-application](../examples/08-application/) – XApp → Deployment + Service.
- **Docs:** [What’s New in v2 – Compose any resource](https://docs.crossplane.io/latest/whats-new/#compose-any-resource), [Compositions – Grant access to composed resources](https://docs.crossplane.io/latest/composition/compositions/#grant-access-to-composed-resources).

---

## Next

- [Composition (XRDs, Compositions, Functions)](./05_Composition_XRDs_Functions.md) – How to define and implement custom APIs.
- [What’s New in v2](./02_Whats_New_In_V2.md) – Namespaced XRs/MRs, compose any resource, Operations.
- Try [examples/05-app-with-deployment](../examples/05-app-with-deployment/) or [examples/08-application](../examples/08-application/) – XApp → Deployment + Service.
