# Official doc: AWS S3 bucket (managed resource)

This example follows the Crossplane guide [Get Started With Managed Resources](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/), which walks through installing an AWS S3 provider and creating a `Bucket` managed resource.

The tutorial uses the **contrib** AWS S3 provider package (`crossplane-contrib/provider-aws-s3`). Installing it also installs **`crossplane-contrib-provider-family-aws`**: the family provider holds shared AWS client/auth wiring; the S3 provider registers the S3 managed resources. Both should show `INSTALLED` and `HEALTHY` before you rely on managed resources.

## Configuration overview

### 1. Provider (`aws-provider.yaml`)

- Declares a `Provider` (`pkg.crossplane.io/v1`) so Crossplane pulls and runs the controller image for that package.
- Package: `xpkg.crossplane.io/crossplane-contrib/provider-aws-s3:v2.0.0` (pin the tag to match the doc or your cluster policy).

### 2. Credentials (not in this repo — create manually)

Crossplane needs AWS credentials in the cluster. The usual pattern from the doc:

1. Put access keys in an INI-style file, for example:

   ```ini
   [default]
   aws_access_key_id = YOUR_KEY_ID
   aws_secret_access_key = YOUR_SECRET
   ```

2. Create a secret in `crossplane-system` whose **key** matches what the provider config references (here: `creds`):

   ```bash
   kubectl create secret generic aws-secret \
     --namespace=crossplane-system \
     --from-file=creds=./aws-credentials.ini
   ```

Other auth methods (IRSA, workload identity, etc.) are covered in the provider and AWS documentation linked from the [same get-started guide](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/); this folder assumes the secret-based flow shown there.

### 3. Cluster provider config (`provider-config.yaml`)

- Kind: `ClusterProviderConfig` (`aws.m.upbound.io/v1beta1`). **Cluster-scoped** — it applies to managed resources in **any** namespace (unlike a namespaced `ProviderConfig`, which only affects resources in that namespace).
- `metadata.name: default` is the common default; managed resources pick this up when they do not set `spec.providerConfigRef` (or when they reference `default` explicitly).
- `spec.credentials` tells the AWS family provider where to read static credentials:
  - `source: Secret`
  - `secretRef`: namespace `crossplane-system`, secret name `aws-secret`, key `creds` (must match the secret you created).

Apply this **after** the provider is healthy **and** the secret exists.

### 4. Bucket (`bucket-test.yaml`)

- API: `s3.aws.m.upbound.io/v1beta1`, kind `Bucket`.
- `spec.forProvider.region` is the AWS region for the bucket (this example uses `eu-central-1`; change as needed).
- S3 bucket names are globally unique. This file uses a fixed `metadata.name` so `kubectl delete -f bucket-test.yaml` targets a stable object; the doc often uses `generateName` instead to avoid name collisions.

## Suggested apply order

1. `kubectl apply -f aws-provider.yaml`
2. Wait until `kubectl get providers` shows the S3 and family providers as healthy.
3. Create the `aws-secret` secret (see above).
4. `kubectl apply -f provider-config.yaml`
5. `kubectl apply -f bucket-test.yaml`

Check readiness: `kubectl get buckets.s3.aws.m.upbound.io` — when `SYNCED` and `READY` are `True`, the bucket exists in AWS.

For prerequisites (cluster size, Crossplane version, AWS account), full command output, and cleanup, see the [official guide](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/).
