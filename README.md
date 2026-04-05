# Crossplane Course (v2)

[![Crossplane docs](https://img.shields.io/badge/docs-docs.crossplane.io-5C4EE5?style=flat-square)](https://docs.crossplane.io/latest/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=flat-square&logo=kubernetes&logoColor=white)](https://kubernetes.io/)

Personal study and course materials for **Crossplane v2**: narrative guides, runnable YAML, and larger **Configurations** you can install, learn from, or use as a base for your own packages. Content is aligned with the [official Crossplane documentation](https://docs.crossplane.io/latest/) and [Crossplane on GitHub](https://github.com/crossplane).

**Repository (replace after you publish):** `https://github.com/YOUR_USERNAME/YOUR_REPO` — fork or clone from your account as needed.

## What this repository is

**Guides (`guides/`)** — The written **course**: concepts, installation, managed resources and providers, composition (XRDs, Compositions, functions), Operations, packages/GitOps, comparisons (e.g. Terraform), Helm → Crossplane, Upbound Marketplace, AWS family providers (contrib vs Upbound), and programmatic YAML. They are numbered and cross-linked from **[Crossplane_Curriculum.md](./Crossplane_Curriculum.md)**, which is the **table of contents and suggested learning path**.

**Examples (`examples/`)** — **Hands-on YAML** that matches and reinforces the guides: minimal S3 buckets and provider config, XRD + Composition pipelines, Patch & Transform, composing Kubernetes workloads (Deployments, Services), CronOperations, Helm `Release` via `provider-helm`, Cue-generated compositions, and the Up CLI workflow. Use the curriculum’s “Practical examples” table to map **topic → folder**.

**Configurations — AWS and beyond**

- **`configuration-aws-eks/`** — Uses an **[Upbound Marketplace](https://marketplace.upbound.io/)** *Configuration* (packaged XRD + Composition + dependencies) to provision **EKS** (and related AWS pieces) through a single composite API. Good for seeing how **curated platform APIs** ship as versioned packages, and for comparing with raw-provider workflows. You can study the pipeline, then **author your own Configuration** (new XRDs, Compositions, `crossplane.yaml`) using the same ideas—see [guides/13_Configurations.md](./guides/13_Configurations.md) and the curriculum’s Packages / Marketplace sections.

- **`eks-cluster-spot/`** — **Provider-based** AWS (VPC, EKS, node groups, etc.) without relying on that Marketplace Configuration: more YAML, more control, useful contrast with `configuration-aws-eks/`.

**`configuration-openforms/`** — A **[Maykin Media Open Forms](https://github.com/maykinmedia/charts/tree/main/charts/openforms)**–style application expressed as **composed Kubernetes resources** (not only cloud managed resources). It shows how Crossplane can drive **in-cluster applications**—Deployments, Services, ConfigMaps, ingress-style concerns—so the control plane is not limited to “cloud buckets and databases.” See that folder’s README and [guides/09_Crossplane_For_Kubernetes_Applications.md](./guides/09_Crossplane_For_Kubernetes_Applications.md).

**`my-platform/`** — A small **custom platform API** (XRD + Composition + package metadata) you can copy or extend when building your own Configuration.

## Start here

- **[Crossplane_Curriculum.md](./Crossplane_Curriculum.md)** — Full index of guides, examples, Configurations, and learning path.
- **[examples/README.md](./examples/README.md)** — Per-example prerequisites and short descriptions.

## Repository layout

```
crossplane-study/
├── README.md                    # This file
├── Crossplane_Curriculum.md     # Course index & learning path
├── guides/                      # Course-style documentation (all topics in curriculum)
├── examples/                    # Runnable examples (tied to guides; see examples/README.md)
│   ├── 00-official-doc-awsbucket/
│   ├── 01-simple-s3-bucket/
│   ├── 02-provider-config-aws/
│   ├── 03-xrd-composition-bucket/
│   ├── 04-patch-and-transform/
│   ├── 05-app-with-deployment/
│   ├── 06-cron-operation/
│   ├── 07-multi-resource-composition/
│   ├── 08-application/
│   ├── configuration-from-helm-chart/
│   ├── cue-xrd-composition/
│   └── up-cli-devex/
├── configuration-aws-eks/       # Upbound Marketplace Configuration (EKS)
├── configuration-openforms/     # App on Kubernetes via Composition (OpenForms-style)
├── eks-cluster-spot/            # Provider-based EKS / VPC (contrast with configuration-aws-eks)
└── my-platform/                 # Minimal custom XRD + Composition sample
```

The [official managed-resources tutorial](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/) manifests (contrib S3 provider, `ClusterProviderConfig`, bucket) live under **`examples/00-official-doc-awsbucket/`**.

## References

- [Crossplane Docs (latest)](https://docs.crossplane.io/latest/)
- [What's New in v2](https://docs.crossplane.io/latest/whats-new/)
- [Crossplane GitHub](https://github.com/crossplane)

## GitHub repository settings (optional)

When you push this repo to **GitHub**, you can mirror common open-source polish:

| Setting | Suggestion |
|--------|------------|
| **About → Description** | Short line from the first paragraph of this README (e.g. “Course-style Crossplane v2 guides, examples, and configurations”). |
| **About → Website** | `https://docs.crossplane.io/latest/` or your own notes URL. |
| **About → Topics** | e.g. `crossplane`, `kubernetes`, `platform-engineering`, `gitops`, `composition`, `infrastructure-as-code`, `xrd` |
| **General → Features** | Enable **Issues** (and **Discussions** if you want Q&A). |
| **General → Wiki** | Off unless you use it. |

Update the **Repository** link near the top of this file to your real `https://github.com/<user>/<repo>` URL.

## Contributing

Issues and pull requests are welcome if you use this as a public repo. For substantive changes, open an issue first so direction matches the course scope. Third-party **Crossplane** and **Kubernetes** trademarks belong to their respective owners; this project is an independent study resource.

## License

This repository is released under the [MIT License](./LICENSE). You may replace it with another license at the repository root if you prefer; update this section accordingly.
