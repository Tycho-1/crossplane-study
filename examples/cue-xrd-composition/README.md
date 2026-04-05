# CUE: Programmatic XRD and Composition Generation

Generate Crossplane XRDs and Compositions from CUE files instead of writing YAML by hand. This example produces the same `XPlatform` XRD and Composition as [my-platform](../my-platform/) and [examples/07-multi-resource-composition](../07-multi-resource-composition/).

## Prerequisites

```bash
# Install CUE
brew install cue
# or: https://github.com/cue-lang/cue/releases
```

## Layout

```
cue-xrd-composition/
├── params.cue      ← Parameter definitions (defaults)
├── values.cue      ← Override params here (edit this file)
├── values-prod.cue← Production overrides (optional)
├── xrd.cue         ← XRD definition (uses params)
├── composition.cue ← Composition definition (uses params)
├── build.sh        ← Script to generate YAML
├── out/            ← Generated output (created by build)
└── README.md
```

## Generate YAML

```bash
./build.sh
```

Outputs:
- `out/xrd.yaml`
- `out/composition.yaml`

### Use different parameters

Edit `values.cue` to change group, kind, region, etc.:

```cue
params: {
	group:          "mycompany.io"
	kind:           "XDataPlatform"
	plural:         "xdataplatforms"
	defaultRegion:  "eu-central-1"
	compositionName: "xplatform"
}
```

### Production build

```bash
ENV=prod ./build.sh
```

Uses `values-prod.cue` (e.g. `defaultRegion: "eu-central-1"`).

### Export manually

```bash
cue export values.cue params.cue xrd.cue -e xrd -o xrd.yaml
cue export values.cue params.cue composition.cue -e composition -o composition.yaml
```

## Validate

```bash
cue vet xrd.cue
cue vet composition.cue
```

## Use the generated manifests

```bash
kubectl apply -f out/xrd.yaml
kubectl apply -f out/composition.yaml
```

Or copy into your Configuration package:

```bash
cp out/xrd.yaml ../my-platform/
cp out/composition.yaml ../my-platform/
```

## Parameters

| Param | Default | Description |
|-------|---------|-------------|
| `group` | `example.org` | API group for the XRD |
| `kind` | `XPlatform` | API kind |
| `plural` | `xplatforms` | Plural name |
| `scope` | `Namespaced` | `Namespaced` or `Cluster` |
| `defaultRegion` | `us-east-2` | Default region for composed resources |
| `compositionName` | `xplatform` | Composition metadata name |

Override in `values.cue` or create `values-<env>.cue` and use `ENV=<env> ./build.sh`.

## Customize

Edit `values.cue` for parameter changes. Edit `xrd.cue` or `composition.cue` for schema changes.

### Adding a new parameter

In `params.cue`:
```cue
#Params: {
	myNewParam: string | *"default"
	// ...
}
```

In `values.cue`:
```cue
params: {
	myNewParam: "custom"
	// ...
}
```

### Adding a new XRD field

In `xrd.cue`, add to `spec.properties` and `required`.

### Adding a new composed resource

In `composition.cue`, add another entry to `resources` in the pipeline input.

## CUE vs KCL vs hand-written YAML

| Approach | When to use |
|----------|-------------|
| **CUE** | Schema validation, reuse, type safety. Good for medium-sized platforms. |
| **KCL** | Complex logic, conditionals, loops. Used in [dot-kubernetes](https://github.com/vfarcic/crossplane-kubernetes). |
| **Hand-written YAML** | Small projects, learning, one-off compositions. |

## See also

- [CUE docs](https://cuelang.org/docs/)
- [CUE Kubernetes tutorial](https://cuelang.org/docs/tutorials/kubernetes/)
- [guides/13_Configurations.md](../../guides/13_Configurations.md) — Configurations overview
- [my-platform/](../../my-platform/) — same XRD/Composition as hand-written YAML
