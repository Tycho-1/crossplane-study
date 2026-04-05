#!/usr/bin/env python3
"""Build apis/composition-openforms.yaml from reference/helm-template-*-external-redis.yaml.

Embeds each rendered object as a P&T base and adds patches for:
  - Secret stringData from XR parameters
  - ConfigMap data keys from XR parameters
  - Deployment replica counts
  - metadata.namespace (fixed default openforms; strings in ConfigMaps assume this namespace)

Helm release name in the reference MUST be \"demo\" so in-cluster strings (ALLOWED_HOSTS,
nginx proxy_pass, etc.) stay consistent. To use another release name, re-render with
helm template MYREL ... and replace \"demo\" in this script / reference, then re-run.

Usage:
  python3 generate_composition.py \\
    ../reference/helm-template-openforms-1.12.0-external-redis.yaml \\
    > ../apis/composition-openforms.generated.yaml
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

import yaml

# Must match helm template --name
RELEASE = "demo"
NS = "openforms"


def split_docs(text: str) -> list[str]:
    parts = re.split(r"^---\s*$", text, flags=re.MULTILINE)
    out: list[str] = []
    for p in parts:
        p = p.strip()
        if not p:
            continue
        # Strip leading comment line (e.g. # Source: chart/templates/...)
        lines = p.splitlines()
        if lines and lines[0].startswith("# Source:"):
            p = "\n".join(lines[1:]).strip()
        if not p:
            continue
        out.append(p)
    return out


def resource_key(doc: dict) -> str:
    return f"{doc.get('kind')}/{doc['metadata']['name']}"


def patches_for(doc: dict) -> list[dict]:
    k = resource_key(doc)
    common_ns = {
        "type": "FromCompositeFieldPath",
        "fromFieldPath": "spec.parameters.namespace",
        "toFieldPath": "metadata.namespace",
    }
    out: list[dict] = [common_ns]

    if k == "Secret/demo-openforms":
        out.extend(
            [
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsSecretKey",
                    "toFieldPath": "stringData.SECRET_KEY",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsDatabasePassword",
                    "toFieldPath": "stringData.DB_PASSWORD",
                },
            ]
        )
    elif k == "ConfigMap/demo-openforms":
        out.extend(
            [
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsBaseUrl",
                    "toFieldPath": "data.BASE_URL",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsDatabaseHost",
                    "toFieldPath": "data.DB_HOST",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsDatabaseName",
                    "toFieldPath": "data.DB_NAME",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsDatabaseUsername",
                    "toFieldPath": "data.DB_USER",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsCeleryBrokerUrl",
                    "toFieldPath": "data.CELERY_BROKER_URL",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsCeleryResultBackend",
                    "toFieldPath": "data.CELERY_RESULT_BACKEND",
                },
                {
                    "type": "FromCompositeFieldPath",
                    "fromFieldPath": "spec.parameters.settingsNumProxies",
                    "toFieldPath": "data.NUM_PROXIES",
                },
            ]
        )
    elif k == "Deployment/demo-openforms":
        out.append(
            {
                "type": "FromCompositeFieldPath",
                "fromFieldPath": "spec.parameters.replicaCount",
                "toFieldPath": "spec.replicas",
            }
        )
    elif k == "Deployment/demo-openforms-worker":
        out.append(
            {
                "type": "FromCompositeFieldPath",
                "fromFieldPath": "spec.parameters.workerReplicaCount",
                "toFieldPath": "spec.replicas",
            }
        )
    elif k == "PersistentVolumeClaim/demo-openforms":
        out.append(
            {
                "type": "FromCompositeFieldPath",
                "fromFieldPath": "spec.parameters.persistenceSize",
                "toFieldPath": "spec.resources.requests.storage",
            }
        )
        out.append(
            {
                "type": "FromCompositeFieldPath",
                "fromFieldPath": "spec.parameters.persistenceStorageClassName",
                "toFieldPath": "spec.storageClassName",
            }
        )
    return out


def build_composition(resources: list[dict]) -> dict:
    return {
        "apiVersion": "apiextensions.crossplane.io/v1",
        "kind": "Composition",
        "metadata": {"name": "openforms-kubernetes"},
        "spec": {
            "compositeTypeRef": {
                "apiVersion": "openforms.example.org/v1alpha1",
                "kind": "OpenForms",
            },
            "mode": "Pipeline",
            "pipeline": [
                {
                    "step": "openforms-resources",
                    "functionRef": {"name": "function-patch-and-transform"},
                    "input": {
                        "apiVersion": "pt.fn.crossplane.io/v1beta1",
                        "kind": "Resources",
                        "resources": resources,
                    },
                },
                {
                    "step": "mark-ready",
                    "functionRef": {"name": "function-auto-ready"},
                },
            ],
        },
    }


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("reference", type=Path)
    ap.add_argument(
        "-o",
        "--output",
        type=Path,
        default=None,
        help="Write full Composition YAML (default: stdout)",
    )
    args = ap.parse_args()
    text = args.reference.read_text()
    docs_raw = split_docs(text)
    resources: list[dict] = []
    for raw in docs_raw:
        doc = yaml.safe_load(raw)
        if not doc or not isinstance(doc, dict):
            continue
        md = doc.get("metadata") or {}
        if not md.get("name"):
            continue
        rk = resource_key(doc)
        safe = rk.replace("/", "-").lower()
        resources.append(
            {
                "name": f"pt-{safe}",
                "base": doc,
                "patches": patches_for(doc),
            }
        )

    comp = build_composition(resources)
    out = yaml.safe_dump(comp, sort_keys=False, default_flow_style=False, width=100)
    if args.output:
        args.output.write_text(out)
    else:
        sys.stdout.write(out)


if __name__ == "__main__":
    main()
