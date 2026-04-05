# EKS Cluster in eu-central-1 with Spot Node Group

Crossplane manifests for an **EKS cluster** in **eu-central-1** and a **managed node group** using **Spot instances**. **Cluster and node IAM roles are defined in Crossplane YAML** and referenced by the Cluster and NodeGroup (no manual ARN copy).

## What’s in this folder

| File | Purpose |
|------|--------|
| **provider-iam.yaml** | Optional: install provider-aws-iam (for IAM Role resources). |
| **provider-eks.yaml** | Optional: install provider-aws-eks (for Cluster and NodeGroup). |
| **provider-config-iam.yaml** | **Required:** Shared ClusterProviderConfig (aws.m.upbound.io). IAM provider has no config CRD; it uses this. |
| **iam-cluster-role.yaml** | IAM Role for EKS cluster (trust `eks.amazonaws.com` + AmazonEKSClusterPolicy). |
| **iam-node-role.yaml** | IAM Role for node group (trust `ec2.amazonaws.com` + EKS worker policies). |
| **cluster.yaml** | EKS Cluster in eu-central-1; uses **roleArnRef** → `eks-cluster-role`. |
| **nodegroup-spot.yaml** | EKS NodeGroup with **capacityType: SPOT**; uses **nodeRoleArnRef** → `eks-node-role`. |

## Prerequisites

1. **Crossplane** installed.
2. **AWS credentials**: a Secret (e.g. `aws-secret` in `crossplane-system`) with your AWS creds; **provider-aws-iam** needs its own ProviderConfig (see **provider-config-iam.yaml**).
3. **VPC and subnets** in eu-central-1 (at least 2 subnets in different AZs). You still need to set **subnet IDs** in the manifests (see below).

## Apply order

1. **Providers** (if not already installed): provider-aws-iam, then provider-aws-eks.
2. **ProviderConfig for IAM**: apply **provider-config-iam.yaml** (so the IAM provider can connect; uses the same Secret as your other AWS config).
3. **IAM roles**: cluster role, then node role.
3. **EKS Cluster** (references cluster role by name).
4. **EKS NodeGroup** (after cluster is READY; references node role by name).

```bash
# 1. Install providers (once)
kubectl apply -f provider-iam.yaml
kubectl apply -f provider-eks.yaml
kubectl get providers   # wait until both are HEALTHY

# 2. ProviderConfig for IAM (Secret aws-secret must exist in crossplane-system)
kubectl apply -f provider-config-iam.yaml

# 3. Create IAM roles (Crossplane creates them in AWS)
kubectl apply -f iam-cluster-role.yaml
kubectl apply -f iam-node-role.yaml
kubectl get role.iam.aws.m.upbound.io   # wait until both are SYNCED / READY

# 4. Edit cluster.yaml: set vpcConfig.subnetIds to your subnet IDs in eu-central-1
kubectl apply -f cluster.yaml

# 5. Wait until cluster is ready (10–15 minutes)
kubectl get cluster.eks.aws.upbound.io my-eks-cluster -w

# 6. Edit nodegroup-spot.yaml: set subnetIds to your subnet IDs (same or subset)
kubectl apply -f nodegroup-spot.yaml
```

## Customize before applying

You only need to set **subnet IDs** (and optionally cluster version, scaling, instance types):

| File | What to set |
|------|-------------|
| **cluster.yaml** | `spec.forProvider.vpcConfig.subnetIds` → your VPC subnet IDs (eu-central-1, at least 2 AZs). |
| **nodegroup-spot.yaml** | `spec.forProvider.subnetIds` → same (or subset of) subnets. |

Cluster and node group **roles are created by Crossplane** and referenced by name (`roleArnRef` / `nodeRoleArnRef`); no ARN copy needed.

Optional:

- **cluster.yaml**: `version` (e.g. `"1.28"`).
- **nodegroup-spot.yaml**: `scalingConfig` (min/max/desired), `instanceTypes`.

## Verify

```bash
# IAM roles
kubectl get role.iam.aws.m.upbound.io

# Cluster
kubectl get cluster.eks.aws.upbound.io
kubectl describe cluster.eks.aws.upbound.io my-eks-cluster

# Node group
kubectl get nodegroup.eks.aws.upbound.io
kubectl describe nodegroup.eks.aws.upbound.io my-eks-spot-nodegroup
```

In AWS: EKS → Clusters → your cluster → Compute → Node groups (capacity type **Spot**).

## API version note

IAM Role manifests use **`iam.aws.m.upbound.io/v1beta1`** so they use the shared **aws.m.upbound.io** ClusterProviderConfig. Cluster and NodeGroup use **`eks.aws.upbound.io/v1beta1`**. All are cluster-scoped unless your providers use namespaced APIs.

## Troubleshooting: Role not created in AWS

The Role is managed by **crossplane-contrib-provider-family-aws** (the one that owns `roles.iam.aws.m.upbound.io`). If the Role exists in the cluster but `aws iam get-role --role-name eks-cluster-role` returns `NoSuchEntity`:

1. **Provider logs** – Check the family provider pod for errors:
   ```bash
   kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=crossplane-contrib-provider-family-aws --tail=100
   ```
   Look for `AccessDenied`, `UnauthorizedOperation`, `InvalidClientTokenId`, or "cannot get ProviderConfig".

2. **AWS IAM permissions** – The credentials in the Secret (used by ClusterProviderConfig `default`) must be allowed to create IAM roles. S3 can work with fewer permissions; IAM needs at least:
   - `iam:CreateRole`
   - `iam:AttachRolePolicy`
   - `iam:GetRole`
   - `iam:DeleteRole` / `iam:DetachRolePolicy` (for delete)
   Attach a policy (e.g. `IAMFullAccess` or a custom policy with these actions) to the IAM user or role whose credentials are in `aws-secret`.

3. **Config and Secret** – Confirm the default config and Secret exist:
   ```bash
   kubectl get clusterproviderconfig.aws.m.upbound.io default
   kubectl get secret aws-secret -n crossplane-system
   ```

4. **Namespace** – The family provider may only reconcile resources in **crossplane-system**. The Role manifests in this folder set `metadata.namespace: crossplane-system` so the provider picks them up. If your Role was in `default`, delete it and re-apply so it is created in `crossplane-system`.

5. **"Timeout: failed waiting for *v1beta1.ClusterProviderConfig Informer to sync"** – The provider’s cache for ClusterProviderConfig never finished syncing. Fix:
   - Ensure the default config exists: `kubectl get clusterproviderconfig.aws.m.upbound.io default`
   - Restart the family provider so it re-syncs: `kubectl delete pod -n crossplane-system -l pkg.crossplane.io/provider=crossplane-contrib-provider-family-aws`
   - Wait for the new pod to be Ready (1–2 min), then check the Role again. If it still fails, the provider’s ServiceAccount may lack permission to list/watch `clusterproviderconfigs.aws.m.upbound.io`; check RBAC for that API group.

6. **Debug logging** – To see reconcile attempts and errors, enable debug on the family provider (edit the Provider to add a `runtimeConfigRef` to a DeploymentRuntimeConfig that adds `--debug` to the container args), then check logs again.

## Cleanup

Delete in reverse order (node group → cluster → IAM roles):

```bash
kubectl delete -f nodegroup-spot.yaml
kubectl delete -f cluster.yaml
# Wait for cluster deletion to finish, then:
kubectl delete -f iam-node-role.yaml
kubectl delete -f iam-cluster-role.yaml
```
