#!/bin/bash
set -e

test -z "$(gofmt -l .)" || { gofmt -l .; exit 1; }
bin="$(go env GOPATH)/bin"
GOBIN="$bin" go install honnef.co/go/tools/cmd/staticcheck@latest
for os in windows linux; do
	GOOS=$os go vet ./...
	GOOS=$os "$bin/staticcheck" ./...
done
