# Programmatic YAML Generation: How Companies Create XRDs and Compositions

Most teams still write YAML by hand. CUE and KCL are used by a minority — mostly larger or more advanced platforms. This guide summarizes what companies actually do and when to use which approach.

---

## What most companies do

### 1. Hand-written YAML (most common)

XRDs and Compositions are edited directly in YAML. No generators, no extra tooling. Works fine for small and medium platforms (a few XRDs, a handful of Compositions). Typical layout: `xrds/` and `compositions/` folders in Git, applied via GitOps.

### 2. KCL (growing, especially for complex platforms)

[KCL](https://kcl-lang.io/) is used by [dot-kubernetes](https://github.com/vfarcic/crossplane-kubernetes) and similar projects. It shines when you need conditionals, loops, and shared logic across many compositions. Python-like syntax, easier for many developers than CUE. Often used to generate large composition YAML (e.g. 200KB+).

### 3. CUE (less common)

CUE offers strong schema and validation, but has a steeper learning curve. It's more common in general Kubernetes tooling (e.g. Kustomize, some internal config stacks) than in Crossplane specifically. Few public Crossplane examples use CUE for XRD/Composition generation.

### 4. Helm (occasional)

Some teams use Helm to template YAML when they're already heavily invested in Helm. No schema validation — you can easily generate invalid YAML. More common for app config than for Crossplane platform definitions.

### 5. Custom generators (Python, Go, etc.)

Internal scripts or tools that emit YAML from a higher-level config. Used by larger platform teams with custom workflows. Rarely shared publicly.

---

## Rough adoption

| Approach | Share of teams | Typical use case |
|----------|----------------|------------------|
| Hand-written YAML | ~70–80% | Small/medium platforms, learning |
| KCL | ~10–15% | Complex, multi-cloud compositions |
| CUE | ~5% | Schema-heavy, validation-focused |
| Helm | ~5% | Teams already templating with Helm |
| Custom scripts | ~5–10% | Large, custom platform tooling |

---

## When to use what

| Situation | Practical choice |
|-----------|------------------|
| Learning, small platform | Hand-written YAML |
| 1–2 XRDs, a few Compositions | Hand-written YAML |
| Many compositions, conditionals, loops | KCL |
| Strong schema/validation needs | CUE |
| Already using Helm everywhere | Helm templating |
| Proprietary platform, many teams | Custom generator (Go/Python) |

---

## Summary

- **Most teams use plain YAML.** It's the default and works well for typical platform sizes.
- **KCL** is the main "programmatic" option in the Crossplane ecosystem — especially for complex, multi-cloud compositions.
- **CUE** is valid but less common for Crossplane; your [cue-xrd-composition example](../examples/cue-xrd-composition/) is a good way to learn parameterization and reuse.
- For production, most teams either stick with YAML or adopt KCL once they outgrow hand-written manifests.

---

## See also

- [examples/cue-xrd-composition](../examples/cue-xrd-composition/) — CUE with parameters
- [guides/13_Configurations.md](./13_Configurations.md) — dot-kubernetes uses KCL
- [vfarcic/crossplane-kubernetes](https://github.com/vfarcic/crossplane-kubernetes) — KCL in practice
