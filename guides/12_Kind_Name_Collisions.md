# Kind Name Collisions: Crossplane + Flux (and other operators)

## The problem

When you run multiple Kubernetes operators in the same cluster, they can each register CRDs that happen to share the same `kind` name. A common example is `Provider`:

| Operator | Full CRD name |
|---|---|
| Crossplane | `providers.pkg.crossplane.io` |
| Flux | `providers.notification.toolkit.fluxcd.io` |

When you use a short kind name in kubectl, kubectl searches all installed CRDs for a match. If two CRDs share the same kind name, kubectl silently picks one — usually alphabetically or by installation order — and gives no warning.

## Real example

With both Crossplane and Flux installed:

```bash
# Ambiguous - kubectl picks Flux's Provider (alphabetically first)
kubectl get provider
# No resources found   <-- correct for Flux, misleading for Crossplane

kubectl explain provider
# GROUP:   notification.toolkit.fluxcd.io   <-- Flux, not Crossplane!
# KIND:    Provider
```

Even `kubectl apply -f aws-provider.yaml` can say `unchanged` (meaning the Crossplane Provider object exists) while `kubectl get provider` returns nothing — because they resolve to different CRDs.

## The fix: always use the full resource name

```bash
# Crossplane providers
kubectl get providers.pkg.crossplane.io

# Flux providers (if needed)
kubectl get providers.notification.toolkit.fluxcd.io
```

## How to detect collisions

```bash
# List all CRDs and look for duplicate kind names
kubectl api-resources | grep -i provider

# General check across all CRDs
kubectl api-resources
```

Look at the `KIND` and `SHORTNAMES` columns — if the same word appears on multiple rows from different API groups, you have a potential collision.

## Why Kubernetes allows this

Kubernetes has no global registry or reservation system for kind names. The API group (`pkg.crossplane.io`, `notification.toolkit.fluxcd.io`) is what makes a resource unique — not the kind name alone. Two operators are free to both define a kind called `Provider`, `Release`, `Repository`, etc., as long as they use different API groups.

The collision only happens at the `kubectl` shorthand level. At the API level, `providers.pkg.crossplane.io` and `providers.notification.toolkit.fluxcd.io` are completely separate and never interfere with each other.

## Other kinds that commonly clash

| Kind | Crossplane | Other operators |
|---|---|---|
| `Provider` | `pkg.crossplane.io` | Flux (`notification.toolkit.fluxcd.io`) |
| `Release` | — | Helm operator, Flux |
| `Repository` | — | Flux, Harbor operator |
| `Bucket` | `s3.aws.m.upbound.io` | Rook-Ceph |

## Recommendation

When working in a cluster with multiple operators, **always use the full `<plural>.<group>` form** for Crossplane resources:

```bash
kubectl get providers.pkg.crossplane.io
kubectl get providerconfigs.aws.crossplane.io
kubectl get compositions.apiextensions.crossplane.io
kubectl get compositeresourcedefinitions.apiextensions.crossplane.io
```

Or check for ambiguity first whenever you get unexpected "no resources found" results:

```bash
kubectl api-resources | grep <kind>
```
