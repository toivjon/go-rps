#!/bin/bash
set -euo pipefail

# Resolve the absolute path of the project root form the script path.
ROOTPATH=$(dirname $(readlink -f $0))
ROOTPATH=$(dirname $ROOTPATH)
cd $ROOTPATH

# Run the unit tests.
printf "Running unit tests. Please wait...\n"
go test -failfast -short -coverprofile coverage.out ./internal/...

# Find the unit tests coverage.
COVERAGE=`go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'`
COVERAGETHRESHOLD=95.0
if (( $(echo "$COVERAGE < $COVERAGETHRESHOLD" | bc -l) )); then
  printf "Checking test coverage failed. Coverage $COVERAGE is less than $COVERAGETHRESHOLD."
  exit 1
fi

# Show information related to test results.
printf "Unit tests succeeded:\n"
printf "    Coverage           $COVERAGE%%\n"
printf "    Coverage Threshold $COVERAGETHRESHOLD%%\n"
printf "Unit tests completed.\n"
