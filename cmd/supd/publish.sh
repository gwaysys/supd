#!/bin/sh
target_dir="publish/supd"
rm -rf $target_dir
mkdir -p $target_dir

export CGO_ENABLED=0

go build -ldflags="-s -w"||exit 0

cp -rf supd $target_dir
cp -rf supc $target_dir 
cp -rf script $target_dir 
cp -rf etc $target_dir
cp -rf setup.sh $target_dir

echo "Publish to $target_dir done"
