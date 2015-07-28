#!/usr/bin/env bash

go get -v -t ./...

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail

set -e

# Set the outfile
outfile=gotest.out

# Run tests
go test -race -v ./... | tee $outfile

# Get go2xunit
which go2xunit > /dev/null || {
    go get bitbucket.org/tebeka/go2xunit
}

# Convert the out file to xml
go2xunit -input $outfile -output tests.xml