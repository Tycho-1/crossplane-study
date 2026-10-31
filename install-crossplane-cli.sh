#!/usr/bin/env bash
# Install the Crossplane CLI (https://docs.crossplane.io/latest/cli/).
# Usage: ./install-crossplane-cli.sh            # pinned version below
#        XP_VERSION=current ./install-crossplane-cli.sh   # latest stable
set -euo pipefail

# v2.4.1 is the newest CLI build on cli.crossplane.io. Later GitHub tags
# (v2.4.2, v2.4.3) are not published there yet, so install.sh cannot fetch them.
export XP_VERSION="${XP_VERSION:-v2.4.1}"

curl -sL "https://raw.githubusercontent.com/crossplane/crossplane/main/install.sh" | sh
sudo mv crossplane /usr/local/bin/

crossplane version --client
