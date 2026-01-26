#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


if [[ "$(dirname $0)" != "." ]]; then
    cd "$(dirname $0)"
fi

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

if [[ ! -f openapi.json ]]; then
    ./get_api_spec.sh
fi

oapi-codegen -config ./cfg.yaml ./openapi.json
