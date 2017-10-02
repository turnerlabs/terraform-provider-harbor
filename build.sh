#!/bin/bash
set -e

go build -v
cp ./terraform-provider-harbor terraform.d/plugins/darwin_amd64
terraform init

rm z.log

# cd test-server
# go build -v