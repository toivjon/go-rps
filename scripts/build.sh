#!/bin/bash
set -euo pipefail

# Resolve the absolute path of the project root form the script path.
ROOTPATH=$(dirname $(readlink -f $0))
ROOTPATH=$(dirname $ROOTPATH)

# Specify the folder where to put build results.
BINPATH="$ROOTPATH/bin/server"

# Clear the old binary folder by re-creating it.
[ -d $BINPATH ] && rm -rf $BINPATH
mkdir -p $BINPATH

# Build the binaries.
printf "Building the binaries. Please wait...\n"
go build -o $BINPATH $ROOTPATH\cmd\server
go build -o $BINPATH $ROOTPATH\cmd\client

# Show information related to compilation.
printf "Build succeeded:\n"
printf "    Server    $BINPATH\server\n"
printf "    Client    $BINPATH\client\n"
printf "Build completed\n"
