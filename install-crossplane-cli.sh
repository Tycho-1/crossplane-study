#!/usr/bin/env bash
# Install the Crossplane CLI (https://docs.crossplane.io/latest/cli/).
# Usage: ./install-crossplane-cli.sh            # pinned version below
#        XP_VERSION=current ./install-crossplane-cli.sh   # latest stable
set -euo pipefail

export XP_VERSION="${XP_VERSION:-v2.4.2}"

curl -sL "https://raw.githubusercontent.com/crossplane/crossplane/main/install.sh" | sh
sudo mv crossplane /usr/local/bin/

crossplane version --client
