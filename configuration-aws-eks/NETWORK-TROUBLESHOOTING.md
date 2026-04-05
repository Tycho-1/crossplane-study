# configuration-aws-network: troubleshooting

This doc covers **CRD / provider API mismatches**, **ProviderConfig vs ClusterProviderConfig**, and **why resources may look “unnamed” in the AWS console**.

---

## The Error ("no matches for kind")

```
failed to get restmapping: no matches for kind "RouteTableAssociation" in version "ec2.aws.m.upbound.io/v1beta1"
failed to get restmapping: no matches for kind "Subnet" in version "ec2.aws.m.upbound.io/v1beta1"
failed to get restmapping: no matches for kind "SecurityGroupRule" in version "ec2.aws.m.upbound.io/v1beta1"
...
```

## Cause

The configuration-aws-network Composition outputs resources with API group **`ec2.aws.m.upbound.io`**. That API group comes from **provider-family-aws**, not from the standalone provider-aws-ec2.

| Provider | API group | Resources |
|----------|-----------|-----------|
| **provider-family-aws** | `ec2.aws.m.upbound.io` | RouteTableAssociation, Subnet, VPC, SecurityGroupRule, etc. |
| **provider-aws-ec2** (standalone, older) | `ec2.aws.upbound.io` | Same resources, different API group |

The Configuration declares `provider-aws-ec2` as a dependency, but the embedded function generates manifests for `ec2.aws.m.upbound.io`. You need the provider that actually provides those CRDs.

---

## Fix: Install provider-family-aws

Install **provider-family-aws** before (or instead of relying solely on) the Configuration's auto-installed provider-aws-ec2:

```yaml
# provider-family-aws.yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: upbound-provider-family-aws
spec:
  package: xpkg.upbound.io/upbound/provider-family-aws:v2.5.0
```

```bash
kubectl apply -f provider-family-aws.yaml
kubectl get providers.pkg.crossplane.io   # wait until INSTALLED=True, HEALTHY=True
```

Then wire AWS credentials. provider-family-aws uses `aws.m.upbound.io`.

**Important:** configuration-aws-network’s Composition sets composed MRs to use **`ProviderConfig`** (namespaced), **not** `ClusterProviderConfig`. So you need a **`ProviderConfig` named `default` in the same namespace as the Network XR** (e.g. `default`). If you only create `ClusterProviderConfig`, you’ll see:

`ProviderConfig.aws.m.upbound.io "default" not found`

Apply [provider-config-default.yaml](./provider-config-default.yaml) (or equivalent). You can still add **ClusterProviderConfig** for other workloads:

```yaml
# Namespaced — required for Network XR in namespace default
apiVersion: aws.m.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
  namespace: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-secret
      key: creds
```

```yaml
# Optional cluster-wide default (other compositions / MRs)
apiVersion: aws.m.upbound.io/v1beta1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: aws-secret
      key: creds
```

See [provider-config-default.yaml](./provider-config-default.yaml) and [cluster-provider-config.yaml](./cluster-provider-config.yaml). The Secret `aws-secret` must exist in `crossplane-system` (key `creds`).

Your Network XR’s `providerConfigName: default` becomes **`ProviderConfig` `default`** in that XR’s namespace (not a lookup of `ClusterProviderConfig` by that field).

---

## Apply Order

1. **provider-family-aws** — install first, wait for HEALTHY
2. **Secret** `aws-secret` in `crossplane-system`
3. **ProviderConfig** `default` in the Network XR namespace ([provider-config-default.yaml](./provider-config-default.yaml)) + optional **ClusterProviderConfig** ([cluster-provider-config.yaml](./cluster-provider-config.yaml))
4. **configuration-aws-network** — install the Configuration, wait HEALTHY
5. **network-xr-example.yaml** (or kustomize Network XR) — create the Network

---

## AWS console names for VPCs and subnets

### What you see in the console

In the **VPC console**, the **Name** column for subnets and VPCs is **not** a separate AWS API field. It is almost always the value of the **`Name` resource tag** (standard AWS tagging). If that tag is missing, the console shows a **blank name** even though the resource exists and has a subnet ID / VPC ID.

So “my subnets have no name” usually means: **no `Name` tag was applied** when the resource was created.

### Where tags come from in this flow

There are three layers; **your YAML only controls the first** unless the platform exposes more.

| Layer | What it does |
|--------|----------------|
| **1. Network XR** (`spec.parameters`) | You choose things like `id`, `region`, `vpcCidrBlock`, and a list of `subnets` (e.g. AZ, `type`, `cidrBlock`). See [network-xr-example.yaml](./network-xr-example.yaml). |
| **2. XRD (CompositeResourceDefinition)** | Defines which fields exist on the **Network** API. If there is **no** field for per-subnet `name` or a generic `tags` map, you **cannot** pass `Name` tags through the XR alone. |
| **3. Composition** (inside **configuration-aws-network**) | The embedded function/KCL decides what each composed **Subnet** / **VPC** managed resource looks like: CIDR, associations, and **which tags** (if any) go into `spec.forProvider` (Terraform AWS provider: `tags`, `tagsAll`, etc.). |

If the Composition **does not** set `Name` (or other display tags) per subnet, **AWS will still create the subnets**; they simply **won’t show a friendly name** in the UI.

### Why network-xr-example.yaml does not set names

The example only specifies:

```yaml
subnets:
  - availabilityZone: eu-central-1a
    type: public
    cidrBlock: 192.168.0.0/18
  # ...
```

There is **no** `name:` or `tags:` here because:

1. The example follows the **minimal** parameters the upstream **Network** API documents for subnets (AZ, type, CIDR).
2. Adding keys that the **XRD does not define** would make `kubectl apply` fail validation.

So missing names in AWS are **expected** for this example unless upstream adds parameters or the Composition sets tags from `id` + subnet index automatically.

### What is set for EKS integration

The important value for **configuration-aws-eks** is the **label** (on Kubernetes composed resources and often mirrored to AWS tags depending on Composition):

`networks.aws.platform.upbound.io/network-id = <your id>`

EKS subnet selection uses that **network id**, not the AWS console **Name** column. So **no Name tag does not break** the EKS + Network integration; it only affects **human readability** in the console.

### What Upbound ships today (verified upstream)

Checked **[upbound/configuration-aws-network](https://github.com/upbound/configuration-aws-network)** (`main`):

| Piece | Subnet / VPC naming |
|--------|---------------------|
| **XRD** [`apis/networks/definition.yaml`](https://github.com/upbound/configuration-aws-network/blob/main/apis/networks/definition.yaml) | Each `subnets[]` item only allows **`availabilityZone`**, **`type`**, **`cidrBlock`**. There is **no** `name`, `tags`, or `mapPublicIpOnLaunch` in the published schema (aside from defaults). |
| **Composition (KCL)** [`functions/network/main.k`](https://github.com/upbound/configuration-aws-network/blob/main/functions/network/main.k) | **VPC** sets `forProvider.tags.Name = oxr.metadata.name` (the **Network XR’s Kubernetes name**). **Subnets** set only `kubernetes.io/role/elb` or `internal-elb` and `networks.aws.platform.upbound.io/network-id` — **no `Name` tag** on subnets. |

So: **Upbound does not support user-defined subnet names (or subnet `Name` tags) via the Network XR API** in that configuration. The VPC may show a name in the AWS console (from the XR `metadata.name`); subnets usually stay **unnamed** in the Name column unless you fork and add tags in KCL.

### How to discover on your cluster (after install)

```bash
kubectl explain network.spec.parameters --recursive
```

This should match the XRD above. If a **newer package version** adds `tags` / `name`, `explain` will show it. Otherwise options are:

- **Upgrade** the Upbound **configuration-aws-network** package if a newer version adds tagging.
- **Fork or wrap** the Configuration and change the Composition to set `tags.Name` (e.g. `"{parameters.id}-public-1a"`).
- **Patch composed MRs** (advanced): edit `Subnet` managed resources’ `spec.forProvider.tags` — fragile, because the next composition reconcile might overwrite them unless the design allows it.

### Kubernetes names vs AWS names

- **`metadata.name`** on the **Network** XR is used by upstream as the **VPC** `Name` tag (`main.k`); it is **not** propagated to **subnet** `Name` tags in the stock Composition.
- Composed **Subnet** MRs use generated Kubernetes resource names; those are unrelated to the EC2 **Name** tag unless the Composition sets `tags.Name` on the subnet.

---

## Adding subnet names (or tags) yourself

**Yes.** Upbound’s package is normal Crossplane + KCL; you change the **schema** (XRD) and the **function** (KCL) that builds each `Subnet`.

### Approach A — Fork (or copy) `configuration-aws-network` (recommended)

1. **Start from upstream**  
   Clone [upbound/configuration-aws-network](https://github.com/upbound/configuration-aws-network) (fork on GitHub if you want to track updates via merge).

2. **Extend the XRD** — `apis/networks/definition.yaml`  
   Under `spec.versions[].schema.openAPIV3Schema.properties.spec.properties.parameters.properties.subnets.items.properties`, add optional fields, for example:

   ```yaml
   # Inside subnets.items.properties (alongside availabilityZone, type, cidrBlock)
   name:
     type: string
     description: Optional AWS Name tag for this subnet (console display name)
   ```

   Or a generic map:

   ```yaml
   tags:
     type: object
     additionalProperties:
       type: string
     description: Extra AWS tags merged onto this subnet (optional)
   ```

   Do **not** remove `required` entries for the existing three fields unless you change defaults accordingly.

3. **Extend the KCL** — `functions/network/main.k`  
   Find the `ec2v1beta1.Subnet{ ... forProvider = { ... tags = { ... } } }` block. Today it only sets the ELB role tags and `network-id`. Merge in a **`Name`** tag and/or user tags, for example:

   - If you added **`name`** on each subnet `s`:

     ```text
     tags = {
       Name = s.name or "{}-{}-{}".format(oxr.spec.parameters.id, s.type, s.availabilityZone)
       # ... existing kubernetes.io/role/* and network-id entries
     }
     ```

   - If you added **`tags`** on each subnet `s`, merge `s.tags` into the same `tags` dict (KCL: build a dict and combine; avoid dropping required keys).

   Regenerate / rebuild the function image as the project already does (see upstream `Makefile` / `upbound.yaml` and CI).

4. **Build and publish your Configuration package**  
   Use **`up project build`** / **`up xpkg push`** (or your registry’s equivalent) so you get an **OCI image** for *your* configuration, e.g. `your-registry.example/cfg-aws-network:v0.1.0`.

5. **Install *your* package on the cluster**  
   Point `configuration-aws-network-install.yaml` (or a copy) at **`spec.package: your-registry.../cfg-aws-network:v0.1.0`** instead of `xpkg.upbound.io/upbound/configuration-aws-network:...`.  
   Wait for the new revision to become **HEALTHY**, then `kubectl explain network.spec.parameters.subnets` should show your new fields.

6. **Use it in YAML** — e.g. [network-xr-example.yaml](./network-xr-example.yaml):

   ```yaml
   subnets:
     - availabilityZone: eu-central-1a
       type: public
       cidrBlock: 192.168.0.0/18
       name: my-public-1a
   ```

**Keeping up with Upbound:** occasionally merge/rebase upstream `configuration-aws-network` into your fork so you pick up bugfixes and provider bumps; resolve conflicts in `definition.yaml` and `main.k`.

### Approach B — No fork: compose `Subnet` MRs yourself

Skip the **Network** XR entirely: define **`Subnet`** (`ec2.aws.m.upbound.io`) resources with full `spec.forProvider.tags` including `Name`. You must also create VPC, routes, IGW, associations, etc., or mix a minimal Network fork with extra MRs.

**Pros:** full control, no XRD change. **Cons:** you re-implement or duplicate what `configuration-aws-network` already wires.

### Approach C — Patch composed MRs in the cluster (usually a bad idea)

`kubectl edit subnet.ec2.aws.m.upbound.io ...` and add `spec.forProvider.tags.Name`. On the next reconcile, **pipeline mode** may **overwrite** desired state from the function output and **drop** your tag unless the Composition emits it. Prefer **A** or **B**.

### Summary

| Approach | Effort | Survives reconcile? |
|----------|--------|---------------------|
| **A — Fork XRD + KCL** | Medium (build/push package) | Yes, if tags are in Composition output |
| **B — Raw MRs** | High (many resources) | Yes |
| **C — Edit MR in cluster** | Low | Often **no** |

---

## Verify CRDs Exist

After installing provider-family-aws:

```bash
kubectl get crd | grep ec2.aws.m.upbound
```

You should see CRDs like:
- `routetableassociations.ec2.aws.m.upbound.io`
- `subnets.ec2.aws.m.upbound.io`
- `vpcs.ec2.aws.m.upbound.io`
- etc.

---

## Alternative: Use eks-cluster-spot VPC (ec2.aws.upbound.io)

If you prefer the standalone provider-aws-ec2 (`ec2.aws.upbound.io`), use the **provider-based** VPC from [eks-cluster-spot](../eks-cluster-spot/):

- `vpc.yaml` — uses `ec2.aws.upbound.io`
- `subnets.yaml` — uses `ec2.aws.upbound.io`

That approach does not use configuration-aws-network. The EKS configuration-aws-eks expects subnets labeled with `networks.aws.platform.upbound.io/network-id` — you'd need to add those labels to the provider-created subnets, or use a different EKS setup.

---

## Summary

| Problem | Solution |
|---------|----------|
| `no matches for kind X in ec2.aws.m.upbound.io` | Install **provider-family-aws** |
| `ProviderConfig.aws.m.upbound.io "default" not found` | Create **namespaced** `ProviderConfig` named `default` in the **Network XR’s namespace** (see [provider-config-default.yaml](./provider-config-default.yaml)); `ClusterProviderConfig` alone does not satisfy that ref |
| Subnets / VPC show **no name** in AWS console | Console **Name** = **`Name` tag**; [network-xr-example.yaml](./network-xr-example.yaml) doesn’t set tags — see [AWS console names for VPCs and subnets](#aws-console-names-for-vpcs-and-subnets); use `kubectl explain network.spec.parameters` or customize the Composition |
| Want **subnet names/tags** on the Network XR | Fork/copy upstream `configuration-aws-network`, extend XRD + `main.k`, build/push OCI, install your package — see [Adding subnet names (or tags) yourself](#adding-subnet-names-or-tags-yourself) |
