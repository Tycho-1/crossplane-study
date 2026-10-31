# Operations

**Operations** in Crossplane v2 are a way to run **declarative, task-based workflows** that run to completion (like a Job) instead of continuously reconciling resources. They use the same **function pipeline** model as Compositions.

**Use cases:**

- Certificate expiry checks and alerts
- Rolling upgrades or maintenance windows
- Scheduled backups or reports
- Reacting to resource changes (e.g. annotate or trigger a job when a resource is updated)

**Status:** Operations are **alpha** in Crossplane v2.

---

## Operation types

| Kind | When it runs |
|------|----------------------|
| **Operation** | Once (on create or when triggered). |
| **CronOperation** | On a schedule (cron expression). |
| **WatchOperation** | When selected resources change. |

All use a **pipeline** of composition functions. The pipeline runs to completion and reports results (e.g. in status or events), rather than managing a set of composed resources forever.

---

## CronOperation example

Run a pipeline on a schedule (e.g. daily at 6 AM):

```yaml
apiVersion: ops.crossplane.io/v1alpha1
kind: CronOperation
metadata:
  name: cert-monitor
spec:
  schedule: "0 6 * * *"   # Cron: daily 6 AM
  operationTemplate:
    spec:
      mode: Pipeline
      pipeline:
        - step: check-certificates
          functionRef:
            name: crossplane-contrib-function-python
      # Function would read certs, check expiry, write result to status or events
```

- **schedule:** Standard cron (e.g. `"0 * * * *"` = every hour).
- **operationTemplate:** Each scheduled run creates an `Operation` from this template.
- **pipeline:** Same idea as Composition: list of steps, each with `functionRef` and optional `input` / `requirements`.

---

## WatchOperation (conceptual)

WatchOperation runs the pipeline when **selected resources** change (e.g. when a specific Deployment or XR is updated). You define:

- Which resources to watch (e.g. by kind, label, namespace).
- The pipeline to run when a change is observed.

(Exact API may differ; check [Operations docs](https://docs.crossplane.io/latest/operations/) for your version.)

---

## Operation vs Composition

| | Composition | Operation |
|---|-------------|-----------|
| Trigger | Create/update XR | Schedule or event (or one-off) |
| Goal | Keep composed resources in sync with XR | Run a task to completion, report result |
| Model | Continuous reconciliation | Run pipeline once per trigger |
| Typical output | Created/updated K8s resources | Status, events, or optional resource changes |

---

## Enabling and using Operations

1. **Feature gate:** Operations are alpha; ensure your Crossplane version supports them and any required flags are set (see [Installation](./03_Installation_Setup.md)).
2. **Functions:** Use the same Function packages as Compositions (e.g. function-python). The function receives a request and returns a response; for Operations the response might indicate “task done” and optional state changes.
3. **RBAC:** The Operation controller needs permission to run the pipeline and, if the pipeline reads/writes resources, appropriate access to those resources.

---

## Practical notes

- Prefer **CronOperation** for periodic checks (certs, backups, reports).
- Use **WatchOperation** when you want to react to specific resource changes.
- Use a single **Operation** for one-off or manually triggered runs.
- Keep pipeline steps idempotent where possible; the same Operation may be retried or run again on the next schedule.

---

## Next

- [Packages & CLI](./07_Packages_CLI.md) – Functions, `crossplane render`.
- [Upgrade & Troubleshooting](./08_Upgrade_Troubleshooting.md).
- Example: [examples/06-cron-operation](../examples/06-cron-operation/).
