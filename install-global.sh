#!/bin/sh
set -e

app=terraform-provider-harbor

version=$(curl -s https://api.github.com/repos/turnerlabs/$app/releases/latest | grep 'tag_name' | cut -d\" -f4)

uname=$(uname)
suffix=""
case $uname in
"Darwin")
suffix="darwin_amd64"
;;
"Linux")
suffix="linux_amd64"
esac

url=https://github.com/turnerlabs/$app/releases/download/$version/$suffix
echo "Getting package $url"

mkdir -p ~/.terraform.d/plugins
curl -sSLo ~/.terraform.d/plugins/$app $url
chmod +x ~/.terraform.d/plugins/$app

echo "harbor terraform plugin installed to ./terraform.d/plugins/$suffix"