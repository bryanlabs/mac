#!/bin/bash
set -e

echo "### Building Windows exe."
#Build and install the Windows executable.
export GOOS="windows"
go build .

echo "### Building and installing Linux Binary."
#Build and install the Linux Binary.
export GOOS="linux"
go build .
go install .