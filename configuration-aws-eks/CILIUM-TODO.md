# Cilium on EKS — TODO for Later

Short reference for when you want to try Cilium instead of the default VPC CNI.

**Clarification:** [Cilium](https://cilium.io/) is a normal Helm chart and widely used on EKS. What is **not** in the [Upbound `configuration-aws-eks`](https://marketplace.upbound.io/configurations/upbound/configuration-aws-eks) package is a **parameter** to pick CNI or to install Cilium for you — the stock composition installs the **VPC CNI** addon path. So “not in Marketplace” here means **not part of that EKS Configuration’s API**, not “Cilium doesn’t exist.” The options below are **sensible** ways to add Cilium anyway: fork + **`provider-helm` `Release`**, use another Configuration that exposes it (**dot-kubernetes**), or **provider-based** full control.

---

## What's needed

- **Disable** VPC CNI addon (and kube-proxy if using Cilium kube-proxy replacement)
- **Install** Cilium via Helm with:
  - `eni.enabled: true`, `ipam.mode: eni` (AWS)
  - `kubeProxyReplacement: strict` (replaces kube-proxy)

---

## configuration-aws-eks does NOT support this today

The package has no parameters for CNI choice or addon toggles. Options:

1. **Fork** configuration-aws-eks → modify KCL to skip vpc-cni addon and add a `Release` (provider-helm) for Cilium
2. **Use dot-kubernetes** — has `spec.parameters.apps.cilium.enabled`
3. **Provider-based** — raw EKS provider + your own Composition with full control

---

## Helm install can be part of a Configuration

- **provider-helm** creates `Release` resources → runs `helm install` in target cluster
- Composition can output a `Release` as a composed resource
- configuration-aws-eks already creates a Helm ProviderConfig (kubeconfig); it just doesn't add any Release
- Add `Release`(Cilium) in KCL after cluster is ready; use `Usage` for ordering

---

## Links

- [Cilium on EKS](https://docs.cilium.io/en/stable/installation/cni-chaining-aws-cni.html)
- [configuration-aws-eks KCL](https://github.com/upbound/configuration-aws-eks/blob/main/functions/eks/main.k)
- [dot-kubernetes](https://marketplace.upbound.io/configurations/devops-toolkit/dot-kubernetes)
