# EKS Cluster in eu-central-1 with Spot Node Group

Crossplane v2 manifests for a complete **EKS** setup in **eu-central-1** using raw **Upbound AWS providers** — no Configuration package. Everything is created by Crossplane: VPC, public subnets, Internet Gateway and routing, IAM roles, the cluster, and a **managed node group on Spot capacity**. Resources find each other by reference (`vpcIdRef`, `subnetIdRefs`, `roleArnRef`, …), so there are no IDs or ARNs to copy.

Compare with [configuration-aws-eks](../configuration-aws-eks/), where one XR creates a similar stack through a curated Marketplace Configuration.

## What’s in this folder

| File | Purpose |
|------|---------|
| **provider-ec2.yaml**, **provider-eks.yaml**, **provider-iam.yaml** | Upbound providers, all pinned to **v2.5.0**. They pull in `upbound-provider-family-aws` automatically. |
| **cluster-provider-config.yaml** | `ClusterProviderConfig` `default` (`aws.m.upbound.io`) — shared credentials for all three providers. |
| **namespace.yaml** | Namespace `eks-spot`; every managed resource below lives here. |
| **vpc.yaml** | VPC `10.0.0.0/16`. |
| **subnets.yaml** | Three public subnets, one per AZ (1a, 1b, 1c). |
| **routing.yaml** | Internet Gateway, public route table with `0.0.0.0/0` route, and one association per subnet. |
| **iam-cluster-role.yaml** | IAM role for the control plane (`AmazonEKSClusterPolicy`). |
| **iam-node-role.yaml** | IAM role for worker nodes (worker, CNI, ECR read-only policies). |
| **cluster.yaml** | EKS cluster, Kubernetes **1.34**, access entries (`authenticationMode: API`). |
| **nodegroup-spot.yaml** | Managed node group, **`capacityType: SPOT`**, AL2023, `t3.medium` + `t3a.medium`. |
| **debug-runtime-config.yaml** | Optional: turn on `--debug` logging for one provider while troubleshooting. |

## Prerequisites

1. **Crossplane v2** installed.
2. **No crossplane-contrib AWS family provider** on the same cluster — it conflicts with the Upbound one. See [guides/18_AWS_Family_Contrib_vs_Upbound.md](../guides/18_AWS_Family_Contrib_vs_Upbound.md).
3. **AWS credentials** in Secret `aws-secret` (key `creds`, AWS INI format) in `crossplane-system`:

   ```bash
   kubectl create secret generic aws-secret -n crossplane-system --from-file=creds=./aws-credentials.ini
   ```

   The IAM identity needs permissions for EC2 networking, IAM roles (`iam:CreateRole`, `iam:AttachRolePolicy`, `iam:PassRole`, …), and EKS.

> **Cost:** this creates a real EKS control plane plus EC2 Spot nodes. Clean up when you are done (see below).

## Apply

```bash
# 1. Providers — wait until all are HEALTHY
kubectl apply -f provider-ec2.yaml -f provider-eks.yaml -f provider-iam.yaml
kubectl get providers.pkg.crossplane.io

# 2. Credentials and namespace
kubectl apply -f cluster-provider-config.yaml -f namespace.yaml

# 3. Network
kubectl apply -f vpc.yaml -f subnets.yaml -f routing.yaml

# 4. IAM roles
kubectl apply -f iam-cluster-role.yaml -f iam-node-role.yaml

# 5. Cluster (10–15 minutes until READY)
kubectl apply -f cluster.yaml
kubectl get clusters.eks.aws.m.upbound.io -n eks-spot -w

# 6. Spot node group
kubectl apply -f nodegroup-spot.yaml
kubectl get nodegroups.eks.aws.m.upbound.io -n eks-spot -w
```

Steps 2–6 can also be applied in one go: references stay unresolved (and resources keep retrying) until their dependencies exist, so Crossplane converges on its own. The order above just makes progress easier to follow.

## Customize

| What | Where |
|------|-------|
| Region / AZs | `region` in every file, `availabilityZone` in **subnets.yaml** |
| CIDRs | **vpc.yaml**, **subnets.yaml** |
| Kubernetes version | `version` in **cluster.yaml** (use a version in EKS standard support) |
| Node sizes and count | `instanceTypes` and `scalingConfig` in **nodegroup-spot.yaml** |
| Names | `metadata.name` is the AWS resource name for the cluster and IAM roles; update references if you rename |

## Verify

```bash
kubectl get managed -n eks-spot     # every resource SYNCED=True, READY=True

# kubeconfig for the new cluster (same IAM identity as aws-secret)
aws eks update-kubeconfig --region eu-central-1 --name my-eks-cluster
kubectl get nodes -L eks.amazonaws.com/capacityType   # expect SPOT
```

`bootstrapClusterCreatorAdminPermissions: true` in **cluster.yaml** gives cluster-admin to the IAM identity Crossplane used to create the cluster. To grant access to another IAM user or role, add an EKS `AccessEntry` + `AccessPolicyAssociation` (the [configuration-aws-eks](../configuration-aws-eks/) `iam.principalArn` parameter does exactly that).

## Troubleshooting

- **Resource stuck at `SYNCED=False`** — `kubectl describe` it; the `Synced` condition message contains the AWS error (e.g. `AccessDenied`, `UnauthorizedOperation`).
- **Reference not resolved** — the referenced object must exist in the **same namespace** (`eks-spot`) and be READY. Check names in `*Ref` fields.
- **`ClusterProviderConfig "default" not found`** — apply **cluster-provider-config.yaml**; check the Secret exists: `kubectl get secret aws-secret -n crossplane-system`.
- **Provider unhealthy with `cannot establish control of object`** — a second AWS family provider (usually crossplane-contrib) is installed. See [guide 18](../guides/18_AWS_Family_Contrib_vs_Upbound.md).
- **Node group fails to create** — check Spot capacity for the instance types in your AZs, and that **routing.yaml** is applied (nodes need internet access to join).
- **Need more detail** — each resource is reconciled by its service provider (Role → `provider-aws-iam`, Cluster/NodeGroup → `provider-aws-eks`, network → `provider-aws-ec2`). Enable debug logs with **debug-runtime-config.yaml** (instructions in the file).

## Cleanup

Delete in reverse order so AWS dependencies are released cleanly:

```bash
kubectl delete -f nodegroup-spot.yaml
kubectl delete -f cluster.yaml
# wait until the cluster is gone (kubectl get clusters.eks.aws.m.upbound.io -n eks-spot)
kubectl delete -f iam-node-role.yaml -f iam-cluster-role.yaml
kubectl delete -f routing.yaml -f subnets.yaml -f vpc.yaml
kubectl delete -f namespace.yaml
```

Keep **cluster-provider-config.yaml** and the providers until every managed resource is deleted — without credentials Crossplane cannot remove the AWS resources.
