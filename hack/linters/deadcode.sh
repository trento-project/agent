#!/usr/bin/env bash

# Copyright 2026 SUSE LLC
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null && pwd)

OUTPUT=$(go tool deadcode -test $PROJECT_DIR/... 2>/dev/null); 
if [ -n "$OUTPUT" ]; then 
    echo "Dead code detected:" >&2; 
    echo "$OUTPUT" >&2; 
    exit 1; 
fi
echo "No dead code found."
