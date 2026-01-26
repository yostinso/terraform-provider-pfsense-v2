#!/usr/bin/env bash

curl -O https://pfrest.org/api-docs/openapi.json
sed -i -e 's|\\\\/|/|g' openapi.json
cp openapi.json openapi_backup.json && jq . openapi_backup.json > openapi.json && rm openapi_backup.json
