#!/usr/bin/env bash

# Copyright 2026 SUSE LLC
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

command -v deadcode >/dev/null 2>&1 || {
    echo "'deadcode' command not found. You can install it locally with 'go install golang.org/x/tools/cmd/deadcode@latest'."
    exit 1
}

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null && pwd)

OUTPUT=$(deadcode -test $PROJECT_DIR/... 2>/dev/null); 
if [ -n "$OUTPUT" ]; then 
    echo "Dead code detected:" >&2; 
    echo "$OUTPUT" >&2; 
    exit 1; 
fi
echo "No dead code found."