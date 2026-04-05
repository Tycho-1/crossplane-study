# Example 06: CronOperation (scheduled pipeline)

**Operations** run a **function pipeline to completion** (like a Job). **CronOperation** runs that pipeline on a **schedule** (cron expression).

**Status:** Operations are **alpha** in Crossplane v2. Ensure your Crossplane version supports them and any required feature flags are set.

This example shows the **shape** of a CronOperation. The pipeline step references a function (e.g. function-python); the function would implement the actual logic (e.g. check certs, run a report).

## Prerequisites

- Crossplane v2 with Operations enabled (alpha)
- A Function that can run in an Operation pipeline (e.g. crossplane-contrib-function-python)

## Apply (if Operations are enabled)

```bash
kubectl apply -f cron-operation.yaml
```

## Verify

```bash
kubectl get cronoperation
kubectl describe cronoperation cert-monitor
```

Check events and status for run history and results.

## Cleanup

```bash
kubectl delete -f cron-operation.yaml
```
