#!/bin/bash
# Generate xrd.yaml and composition.yaml from CUE sources
# Use ENV=prod to pick values-prod.cue (default: values.cue)
set -e

OUT_DIR="${OUT_DIR:-./out}"
ENV="${ENV:-}"
mkdir -p "$OUT_DIR"

# Pick values file: values-prod.cue if ENV=prod, else values.cue
if [ "$ENV" = "prod" ]; then
  VALUES="values-prod.cue"
  echo "Using $VALUES (ENV=prod)"
else
  VALUES="values.cue"
fi

echo "Generating xrd.yaml..."
cue export "$VALUES" params.cue xrd.cue -e xrd -o "$OUT_DIR/xrd.yaml"

echo "Generating composition.yaml..."
cue export "$VALUES" params.cue composition.cue -e composition -o "$OUT_DIR/composition.yaml"

echo "Done. Output in $OUT_DIR/"
ls -la "$OUT_DIR"
