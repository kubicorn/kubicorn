#!/usr/bin/env bash
set -e
cd ~

# VPN Mesh
curl http://meshbird.com/install.sh | sh
export MESHBIRD_KEY=$MESHBIRD_KEY
meshbird join &
