# Crossplane Course – Practical Examples

All examples are **runnable** with a Crossplane cluster and the listed prerequisites. Apply order (XRD → Composition → XR) matters where noted.

| Example | Description | Prerequisites |
|---------|-------------|---------------|
| [00-official-doc-awsbucket](./00-official-doc-awsbucket/) | **Official doc parity:** contrib `provider-aws-s3`, `ClusterProviderConfig`, `Bucket` ([get-started](https://docs.crossplane.io/latest/get-started/get-started-with-managed-resources/)) | Crossplane, AWS secret + family/S3 providers healthy |
| [up-cli-devex](./up-cli-devex/) | Up CLI workflow: project init, xrd generate, composition generate | up CLI |
| [01-simple-s3-bucket](./01-simple-s3-bucket/) | Single S3 Bucket (managed resource) | provider-aws-s3, ProviderConfig |
| [02-provider-config-aws](./02-provider-config-aws/) | ClusterProviderConfig + ProviderConfig for AWS | AWS credentials, Secret |
| [03-xrd-composition-bucket](./03-xrd-composition-bucket/) | Custom API (XBucket) → S3 Bucket | provider-aws-s3, function-patch-and-transform |
| [04-patch-and-transform](./04-patch-and-transform/) | P&T patches (region, labels), `crossplane render` | Same as 03 |
| [05-app-with-deployment](./05-app-with-deployment/) | **v2:** XApp → Deployment + Service (any K8s resource) | function-patch-and-transform, RBAC |
| [06-cron-operation](./06-cron-operation/) | CronOperation (alpha) – scheduled pipeline | Operations enabled |
| [07-multi-resource-composition](./07-multi-resource-composition/) | One XR → multiple composed resources (Bucket) | Same as 03 |
| [08-application](./08-application/) | **Application:** Kubernetes app (XApp → Deployment + Service), no cloud | function-patch-and-transform, RBAC |
| [configuration-from-helm-chart](./configuration-from-helm-chart/) | Helm `Release` via `provider-helm` (sample) | provider-helm + ProviderConfig; see [guide 19](../guides/19_Configuration_From_Helm_Chart.md) |
| [cue-xrd-composition](./cue-xrd-composition/) | Generate XRD/Composition YAML from Cue (`build.sh` → `out/`) | cue CLI (optional), Crossplane for apply |

## Application examples (v2)

Crossplane can create **normal Kubernetes applications** (not only cloud resources):

- **[05-app-with-deployment](./05-app-with-deployment/)** – XApp → Deployment + Service.
- **[08-application](./08-application/)** – Same idea; dedicated **application** example (XApp with image, replicas, port → Deployment + Service).

See [guides/09_Crossplane_For_Kubernetes_Applications.md](../guides/09_Crossplane_For_Kubernetes_Applications.md). For **Helm `Release`** via Crossplane, see [guides/19_Configuration_From_Helm_Chart.md](../guides/19_Configuration_From_Helm_Chart.md) and [configuration-from-helm-chart](./configuration-from-helm-chart/). For **Open Forms** as composed K8s resources (no Helm at runtime), see **[configuration-openforms](../configuration-openforms/)**.

## Quick start

1. Install Crossplane and provider-aws-s3 ([guides/03_Installation_Setup.md](../guides/03_Installation_Setup.md)).
2. Optionally follow the official tutorial verbatim ([00-official-doc-awsbucket](./00-official-doc-awsbucket/)); or create AWS credentials Secret and ClusterProviderConfig ([02-provider-config-aws](./02-provider-config-aws/)).
3. Create a bucket via MR ([01-simple-s3-bucket](./01-simple-s3-bucket/)).
4. Define a custom API and Composition ([03-xrd-composition-bucket](./03-xrd-composition-bucket/)).

See [Crossplane_Curriculum.md](../Crossplane_Curriculum.md) for the full course index.
