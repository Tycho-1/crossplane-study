# Example 02: Provider Config (AWS credentials)

Configure AWS credentials for the Upbound AWS provider using **ClusterProviderConfig** and a Kubernetes Secret.

## 1. Create the secret

**Option A – from AWS credentials file:**

```bash
kubectl create secret generic aws-secret \
  -n crossplane-system \
  --from-file=creds=$HOME/.aws/credentials
```

**Option B – from environment (access key + secret):**

```bash
kubectl create secret generic aws-secret \
  -n crossplane-system \
  --from-literal=creds="[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
"
```

## 2. Apply the ClusterProviderConfig

```bash
kubectl apply -f cluster-provider-config.yaml
```

This creates a **default** ClusterProviderConfig so all MRs (in any namespace) can use it without setting `providerConfigRef`.

## 3. Optional: namespaced ProviderConfig

If you want a **namespaced** config (e.g. for a specific team namespace):

```bash
kubectl apply -f provider-config.yaml
```

Then in MRs in that namespace, set:

```yaml
spec:
  providerConfigRef:
    name: my-aws-config
    kind: ProviderConfig
```

## Verify

After applying, create a Bucket MR (e.g. from [01-simple-s3-bucket](../01-simple-s3-bucket/)). It should become `SYNCED` and `READY` if credentials are valid.
