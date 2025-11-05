#!/bin/bash
# Script to fix util.go imports

cd "$(dirname "$0")"

# Replace e2e.Logf with exutil.Logf
sed -i 's/e2e\.Logf/exutil.Logf/g' util.go

# Replace e2eoutput.RunHostCmd with exutil.RunHostCmd
sed -i 's/e2eoutput\.RunHostCmd/exutil.RunHostCmd/g' util.go

echo "Fixed util.go imports"

