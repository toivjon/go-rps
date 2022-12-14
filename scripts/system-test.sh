#!/bin/bash
set -euo pipefail

# Resolve the absolute path of the project root form the script path.
ROOTPATH=$(dirname $(readlink -f $0))
ROOTPATH=$(dirname $ROOTPATH)
cd $ROOTPATH

# Build the binaries to get the latest version of the applications.
./scripts/build.sh

# Run the system tests.
printf "Running system tests. Please wait...\n"
go run ./systest/client
go run ./systest/server

# Show information related to test results.
printf "System tests succeeded\n"
