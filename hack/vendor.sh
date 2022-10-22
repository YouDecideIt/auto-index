#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# We cannot use the directory ./vendor as it's used by go mod vendor
rm -rf replaced
pushd / > /dev/null
cache=$(go mod download -json k8s.io/kube-openapi@v0.0.0-20210421082810-95288971da7e  | jq -r '.Dir')
popd > /dev/null
destp=replaced/k8s.io/kube-openapi@v0.0.0-20210421082810-95288971da7e
for f in $(find $cache); do
    relative=${f#"$cache"}
    if [ -d $f ]; then
        mkdir -p $destp/$relative
    else
        mkdir -p $(dirname $destp/$relative)
        cp $f $destp/$relative
    fi
done
git apply hack/vendor.patch
