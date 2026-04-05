# Reference `helm template` output

These files are **not** applied by Crossplane. They are **snapshots** from the upstream Helm chart so you can **diff** when upgrading and **regenerate** [apis/composition.yaml](../apis/composition.yaml) with [scripts/generate_composition.py](../scripts/generate_composition.py).

| File | Contents |
|------|----------|
| **helm-template-openforms-1.12.0-external-redis.yaml** | Chart **1.12.0**, **Redis subchart disabled**, external broker/result URLs — **this** is what the current Composition embeds. |
| **helm-template-openforms-1.12.0-full.yaml** | Same chart with **default** embedded Redis (Bitnami) — use when porting StatefulSet/Services for in-cluster Redis. |
