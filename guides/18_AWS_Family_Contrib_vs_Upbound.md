# AWS on Crossplane: contrib vs Upbound — why mixing breaks

This guide explains why **crossplane-contrib** and **Upbound** AWS **family** providers conflict, what error you see, and how to run a clean cluster. It complements [Managed Resources & Providers](./04_Managed_Resources_Providers.md), [Packages & CLI](./07_Packages_CLI.md), and [configuration-aws-eks](../configuration-aws-eks/).

---

## Two product lines (not interchangeable on one cluster)

| Line | Typical package registry | Examples |
|------|--------------------------|----------|
| **Upbound official** | `xpkg.upbound.io/upbound/...` | `provider-family-aws`, `provider-aws-ec2`, `provider-aws-eks`, Marketplace Configurations |
| **Community (contrib)** | `xpkg.crossplane.io/crossplane-contrib/...` | `provider-family-aws`, `provider-aws-s3`, etc. |

Both implement **Crossplane AWS providers** (often built from similar Upjet codebases), but they ship as **different OCI packages** with **different package controller identity**.

---

## Why the conflict is “so big”

**They are not two independent plugins.** Both **family** packages install and claim ownership of the **same Kubernetes APIs**, including:

- `aws.m.upbound.io` — e.g. `ProviderConfig`, `ClusterProviderConfig`, **`ProviderConfigUsage`**
- The same **CRDs** and Crossplane’s rule that **one active `ProviderRevision`** establishes **control** of those types

So:

1. **contrib `provider-family-aws`** installs and registers as controller for `providerconfigusages.aws.upbound.io` (and related plumbing).
2. **Upbound `provider-family-aws`** needs the **same** control for its revision to be healthy.

Crossplane allows **only one** revision to own that control → the second reports **UnhealthyPackageRevision**, e.g.:

```text
cannot establish control of object: providerconfigusages.aws.upbound.io is already controlled by ProviderRevision crossplane-contrib-provider-family-aws-...
```

This is **by design**, not a random bug: **one family controller per cluster** for that API surface.

---

## Many AWS service providers are fine — contrib vs Upbound is the problem

The limit is **not** “only one AWS-related `Provider` on the cluster.”

| Situation | OK? |
|-----------|-----|
| **Several service packages, same line (e.g. all Upbound)** — `provider-family-aws` + `provider-aws-ec2` + `provider-aws-eks` + `provider-aws-s3` + `provider-aws-iam` | **Yes** — normal and expected. |
| **Two different stacks** — **contrib** family vs **Upbound** family (or mixing family from one line with services from the other without support) | **No** — they are **different product lines** that fight over the same cluster-scoped family APIs. |

So: **multiple service providers = fine.** The breakage comes from installing **contrib** and **Upbound** as **competing** AWS family implementations, not from having “too many” Upbound EC2/EKS/S3 packages.

---

## Should contrib and Upbound “just work” together?

**No** — not for **two family providers** at once.

- **Do not** install **both** `crossplane-contrib-provider-family-aws` and `upbound-provider-family-aws`.
- **Avoid** mixing **contrib family** with **Upbound service** providers (or the reverse) unless upstream docs explicitly support it — version and ownership assumptions differ.

**Recommended:** Pick **one** line for AWS and stick to it (especially for EKS / EC2 / Network Configurations from the Upbound Marketplace → use **Upbound** packages).

---

## What to run (Upbound-only example)

1. Remove contrib family (and migrate off contrib-only packages if needed):

   ```bash
   kubectl delete provider.pkg.crossplane.io crossplane-contrib-provider-family-aws
   ```

2. Install **one** family:

   ```bash
   kubectl apply -f configuration-aws-eks/provider-family-aws.yaml   # upbound provider-family-aws
   ```

3. Let **Configurations** pull `provider-aws-ec2` / `provider-aws-eks` via `dependsOn`, or install those Providers explicitly with **matching** family versions.

4. Verify:

   ```bash
   kubectl get providers.pkg.crossplane.io
   ```

   Expect **HEALTHY True** for family + EC2/EKS as applicable.

---

## After delete, errors still mention “contrib”

Deleting the **`Provider`** does not always remove stale **`ProviderRevision`** objects. The old revision can still hold **control** until removed.

```bash
kubectl get providerrevisions.pkg.crossplane.io | grep -i contrib
kubectl delete providerrevisions.pkg.crossplane.io <stale-contrib-revision-name>
```

If delete hangs on **finalizers** (only if you are sure the revision is orphaned):

```bash
kubectl patch providerrevisions.pkg.crossplane.io <name> \
  -p '{"metadata":{"finalizers":[]}}' --type=merge
```

Then delete and re-apply the Upbound family Provider:

```bash
kubectl delete provider.pkg.crossplane.io upbound-provider-family-aws
kubectl apply -f configuration-aws-eks/provider-family-aws.yaml
```

Optional: restart Crossplane if the API server still shows stale state:

```bash
kubectl rollout restart deployment -n crossplane-system -l app=crossplane
# adjust label/name to match your install
```

---

## Related errors (same theme)

| Symptom | Likely cause |
|---------|----------------|
| `providerconfigs ... forbidden` for `upbound-provider-aws-ec2` | RBAC from package not applied; reinstall **provider-aws-ec2** after family is healthy; align versions |
| `no matches for kind ... ec2.aws.m.upbound.io` | EC2 CRDs not installed; need **provider-aws-ec2** (often via Configuration `dependsOn`) |
| Two `provider-family-aws` rows | Remove one; see above |

---

## See also

- [configuration-aws-eks README](../configuration-aws-eks/README.md) — EKS/Network manifests, [provider-family-aws.yaml](../configuration-aws-eks/provider-family-aws.yaml)
- [NETWORK-TROUBLESHOOTING.md](../configuration-aws-eks/NETWORK-TROUBLESHOOTING.md) — VPC / API group issues
- [Provider packages and Kubernetes RBAC](./13_Configurations.md#provider-packages-and-kubernetes-rbac) — who creates ClusterRole / bindings
- [Upbound provider families (docs)](https://docs.upbound.io/providers/provider-families/)
- [Upbound Marketplace](https://marketplace.upbound.io/providers) — official package names and versions
