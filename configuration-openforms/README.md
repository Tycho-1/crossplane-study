# configuration-openforms

Crossplane **Configuration** (and local manifests) that deploy **[Open Forms](https://open-forms.readthedocs.io/)** using **composed Kubernetes resources only** — **no** `provider-helm` and **no** Helm `Release` at runtime. Manifests were derived from the **[maykinmedia/openforms](https://github.com/maykinmedia/charts/tree/main/charts/openforms)** Helm chart **v1.12.0** (app **3.3.9**).

**→ End users / marketplace:** parameter reference, quickstart, and Helm-value mapping — **[USAGE.md](./USAGE.md)** (similar role to the [upstream chart README](https://github.com/maykinmedia/charts/blob/main/charts/openforms/README.md)).

This folder mirrors the role of **[configuration-aws-eks](../configuration-aws-eks/)**: a dedicated place for install docs, examples, and reference material — here for an **application** XR instead of EKS.

---

## What you get

| Piece | Role |
|-------|------|
| **[apis/definition.yaml](./apis/definition.yaml)** | XRD: `OpenForms` (`openforms.example.org/v1alpha1`) |
| **[apis/composition.yaml](./apis/composition.yaml)** | Pipeline: **Patch & Transform** (embedded chart manifests) + **function-auto-ready** |
| **[crossplane.yaml](./crossplane.yaml)** | Package metadata + `dependsOn` for P&T and auto-ready |
| **[reference/](./reference/)** | `helm template` snapshots for **diffing** when the upstream chart changes |
| **[scripts/generate_composition.py](./scripts/generate_composition.py)** | Regenerates `apis/composition.yaml` from the external-Redis reference |

**Scope of the port (important):**

- The **Composition** embeds the **external Redis** variant: `helm template` with **`tags.redis=false`** and broker/result URLs from the XR. You must run **Redis** (or compatible) yourself.
- A **full** render **with** the Bitnami Redis subchart is kept as **[reference/helm-template-openforms-1.12.0-full.yaml](./reference/helm-template-openforms-1.12.0-full.yaml)**. To ship Redis **inside** the same Composition, copy the extra objects from that file into `composition.yaml` (or extend the generator) and add parameters — same maintenance story.

---

## Hard constraint: XR name `demo`

Embedded objects use the Helm **release name** **`demo`** in `metadata.name` (`demo-openforms`, …), **labels** (`app.kubernetes.io/instance: demo`), and **ConfigMap** strings (`ALLOWED_HOSTS`, nginx `proxy_pass`, rate-limit zones). **Your OpenForms XR `metadata.name` must be `demo`.**

To use another release name, re-render with `helm template MYREL ...`, replace `MYREL` consistently in the reference YAML, and run `scripts/generate_composition.py` again (or sed the bases).

---

## Namespace and embedded DNS strings

`spec.parameters.namespace` patches **`metadata.namespace`** on every object (default **`openforms`**). Internal strings such as **`demo-openforms.openforms`** in ConfigMaps assume **namespace `openforms`**. If you change the namespace parameter, regenerate from Helm with `-n <ns>` and update the reference, then re-run the generator — **do not** only patch `metadata.namespace`.

---

## Prerequisites

- Crossplane **v2** with **Composition functions** (Pipeline mode).
- **function-patch-and-transform** and **function-auto-ready** installed (see [examples/05-app-with-deployment](../examples/05-app-with-deployment/functions.yaml)).
- **RBAC** allowing Crossplane to create the composed types (ServiceAccount, Secret, ConfigMap, PVC, Service, Deployment, etc.) in the target namespace — same idea as [05-app-with-deployment README](../examples/05-app-with-deployment/README.md).

---

## Install (without building an OCI package)

```bash
kubectl apply -f apis/definition.yaml
kubectl apply -f apis/composition.yaml
# Install functions if needed, then:
kubectl apply -f openforms-xr-example.yaml
```

Edit **[openforms-xr-example.yaml](./openforms-xr-example.yaml)** (secrets, DB, Redis URLs) before applying.

---

## Build and install as a Configuration package

From this directory:

```bash
# Exclude reference helm dumps, example install CR, generator script, and sample XR — they are not package metadata.
crossplane xpkg build --package-root=. -o configuration-openforms.xpkg \
  --ignore='reference/*.yaml,examples/*.yaml,scripts/*.py,openforms-xr-example.yaml'
# push to your OCI registry, then apply the install manifest:
kubectl apply -f examples/install-from-registry.yaml
```

Adjust **`YOUR_REGISTRY/configuration-openforms:v0.1.0`** in [examples/install-from-registry.yaml](./examples/install-from-registry.yaml).

---

## Updating when the upstream chart changes

1. Pull the new chart version and run **`helm template`** with the same options used for [reference/helm-template-openforms-1.12.0-external-redis.yaml](./reference/helm-template-openforms-1.12.0-external-redis.yaml) (document flags in a comment at the top of that file when you change them).
2. Diff the new YAML against the saved reference.
3. Run:

   ```bash
   python3 scripts/generate_composition.py \
     reference/helm-template-openforms-1.12.0-external-redis.yaml \
     -o apis/composition.yaml
   ```

4. Re-add any **hand-written** patches in `generate_composition.py` if you extended `patches_for()`.
5. Rebuild the **xpkg** and roll out.

---

## Relationship to Helm

- **Helm is a development tool** here: it renders the chart to **reference** YAML and keeps the port honest.
- **Runtime** is **Crossplane Composition** only — no Helm operator on the cluster for this app.

**Design note:** Why we expose only **some** chart values as XR parameters — and when **`provider-helm` `Release`** might be a better fit than mirroring all of `values.yaml` — is explained in **[USAGE.md — Design: curated parameters vs full Helm values](./USAGE.md#design-curated-parameters-vs-full-helm-values)**.

---

## Chart reference

- Upstream chart: [maykinmedia/charts — openforms](https://github.com/maykinmedia/charts/tree/main/charts/openforms)  
- Chart repo: `https://maykinmedia.github.io/charts/`
