#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


if [[ "$(dirname $0)" != "." ]]; then
    cd "$(dirname $0)"
fi

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

if [[ ! -f openapi.json ]]; then
    curl -O https://pfrest.org/api-docs/openapi.json
    sed -i -e 's|\\\\/|/|g' openapi.json
    cp openapi.json openapi_backup.json && jq . openapi_backup.json > openapi.json && rm openapi_backup.json
fi

oapi-codegen -config ./cfg.yaml ./openapi.json
