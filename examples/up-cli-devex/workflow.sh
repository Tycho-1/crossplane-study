#!/bin/bash
# Example workflow: create a new Crossplane project with Up CLI
# Run step-by-step or as reference. Requires: up CLI, Docker

set -e

PROJECT_NAME="${1:-my-s3-platform}"
echo "Creating project: $PROJECT_NAME"

# 1. Initialize project
up project init "$PROJECT_NAME"
cd "$PROJECT_NAME"

# 2. Add dependencies
up dependency add "xpkg.upbound.io/upbound/provider-aws-s3@>=v1.22.0"
up dependency add "xpkg.upbound.io/crossplane-contrib/function-auto-ready@>=v0.0.0"

# 3. Generate example (XRC = claim)
up example generate --name my-bucket \
    --namespace default \
    --type xrc \
    --kind StorageBucket \
    --api-group platform.example.com \
    --api-version v1alpha1

# 4. Edit examples/.../my-bucket.yaml with your desired spec, then:
# up xrd generate examples/storagebucket/my-bucket.yaml

# 5. Generate composition
# up composition generate apis/xstoragebuckets/definition.yaml --path xstoragebuckets/aws-s3.yaml

# 6. Generate function
# up function generate aws-s3 apis/xstoragebuckets/aws-s3.yaml

# 7. Build
# up project build

# 8. Push (requires registry auth)
# up project push --tag v0.1.0

echo "Project created. Next: edit examples, then run up xrd generate, up composition generate, up function generate, up project build, up project push"
