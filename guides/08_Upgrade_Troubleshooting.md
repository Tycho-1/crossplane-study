# Upgrade & Troubleshooting

## Upgrading to Crossplane v2

### Before upgrading

1. **Fully qualified package URLs**  
   v2 requires every package to use a full image URL (e.g. `xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0`). Check:
   ```bash
   kubectl get pkg
   ```
   Fix any package that doesn’t use a full URL.

2. **Deprecated/removed features**  
   If you use any of these, plan a migration before upgrading:
   - **Native P&T in Compositions** → migrate to **function-patch-and-transform** (or another function).
   - **ControllerConfig** → use **DeploymentRuntimeConfig** for Provider/Function runtime.
   - **External secret stores** → use standard Secrets + ProviderConfig.
   - **Composite connection details** → compose your own Secret (see [connection details composition](https://docs.crossplane.io/latest/guides/connection-details-composition/)).

3. **Backup**  
   Backup CRs (XRDs, Compositions, XRs, MRs) and, if needed, export cloud resource state.

### Upgrade steps

1. Follow the official [Upgrade to Crossplane v2](https://docs.crossplane.io/latest/guides/upgrade-to-crossplane-v2/) guide.
2. Upgrade the Crossplane Helm release (or manifest) to a v2 version.
3. Upgrade providers and functions to v2-compatible versions.
4. Update Compositions to use **Pipeline** mode and functions; update XRDs to v2 API and `scope: Namespaced` if desired.
5. Migrate existing XRs/MRs to v2-style (namespaced) if you adopt the new model; the guide covers compatibility and migration.

### After upgrading

- Confirm `kubectl get providers` and `kubectl get functions` show `HEALTHY: True`.
- Confirm XRs reconcile and composed resources are created/updated.
- Check for deprecation warnings in events or provider/function status.

---

## Troubleshooting

### Provider not healthy

- **Symptoms:** `kubectl get providers` shows `HEALTHY: False` or `Unknown`.
- **Checks:**
  - `kubectl describe providerrevisions` – events often show “incompatible Crossplane version”, missing dependencies, or image pull errors.
  - Dependencies: ensure any required Configurations or other Providers are installed and healthy.
  - Image: ensure `spec.package` is correct and pullable (network, imagePullSecrets if private).

### MR stuck (not syncing / not ready)

- **Symptoms:** MR stays `SYNCED: False` or `READY: False`.
- **Checks:**
  - `kubectl describe <mr-kind> <name>` – conditions and events (e.g. auth failure, quota, invalid spec).
  - **ProviderConfig:** correct secret ref, namespace, and key; secret present and with valid cloud credentials.
  - Cloud side: check cloud console/CLI for errors (permissions, limits, region).

### XR not composing resources

- **Symptoms:** XR exists but no composed resources appear, or pipeline doesn’t run.
- **Checks:**
  - **Composition selection:** XR’s `spec.crossplane.compositionRef` or default – is there a Composition whose `compositeTypeRef` matches this XR’s kind/version?
  - **Functions:** `kubectl get functions` – are all referenced functions `INSTALLED` and `HEALTHY`?
  - **CompositionRevision:** `kubectl get compositionrevision` – is there a revision for the Composition? Try `kubectl describe xr <name>` and look at “Composition Ref” / “Resource Refs”.
  - **RBAC:** If composing non-Crossplane resources, does Crossplane have RBAC (aggregate ClusterRole) for those types?
  - **Render locally:** `crossplane render xr.yaml composition.yaml functions.yaml` to see if the pipeline produces resources and to debug function input/output.

### Operation (CronOperation / WatchOperation) not running

- **Feature:** Operations are alpha; ensure your Crossplane version and feature flags support them.
  - **Schedule:** For CronOperation, confirm cron expression and timezone.
  - **Pipeline:** Same as Compositions – functions must be installed and healthy; check events on the Operation resource.

### Deleting when a resource hangs

- **MR stuck in Terminating:** Provider may not be able to delete the cloud resource (e.g. finalizers, dependency). Options (with care):
  - Fix the cause (e.g. delete dependency in cloud first, fix provider).
  - Remove finalizers from the MR so the object is removed from the cluster (cloud resource may remain and need manual cleanup). See [Troubleshooting guide](https://docs.crossplane.io/latest/guides/troubleshoot-crossplane/#deleting-when-a-resource-hangs).

### Package install slow or API server load

- **Many CRDs:** Providers can install hundreds of CRDs. Consider **Managed Resource Definitions (MRDs)** to activate only the MRs you need (v2).
  - **Resource sizing:** Ensure API server and etcd have enough resources; see [scaling CRDs](https://github.com/crossplane/crossplane/blob/main/design/one-pager-crd-scaling.md).

---

## Useful commands

```bash
# Crossplane and packages
kubectl get pods -n crossplane-system
kubectl get providers
kubectl get functions
kubectl get providerrevisions
kubectl describe providerrevisions <name>

# Compositions and XRs
kubectl get composition
kubectl get compositionrevision
kubectl get composite
kubectl get xrd

# Managed resources (example)
kubectl get bucket -A
kubectl describe bucket <name> -n <ns>
```

---

## References

- [Troubleshoot Crossplane](https://docs.crossplane.io/latest/guides/troubleshoot-crossplane/)
- [Upgrade to Crossplane v2](https://docs.crossplane.io/latest/guides/upgrade-to-crossplane-v2/)
- [Deleting when a resource hangs](https://docs.crossplane.io/latest/guides/troubleshoot-crossplane/#deleting-when-a-resource-hangs)
