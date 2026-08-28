#!/bin/sh
target_dir="publish/supd"
rm -rf $target_dir
mkdir -p $target_dir

export CGO_ENABLED=0

go build -ldflags="-s -w" -o bin/supd cmd/supd/main.go ||exit 0

# The setup logic now lives inside the supd binary itself (see setup.go).
# Final deliverable is a single binary + etc configuration; no shell wrappers shipped.

cp -rf bin $target_dir
cp -rf README.md $target_dir

echo "Publish to $target_dir done"
