#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


function get_for_tag() {
    jq '
        .paths | with_entries(select(
            .value | with_entries(select(
                .value.tags[]? | index("'$1'")
            )) | length > 0
        )) | keys
    ' $(dirname $0)/openapi.json
}

if [[ -z "$1" ]]; then
    (
        echo "{"
        (
            tags=$(jq -r '.tags | map(.name) | join(" ")' openapi.json)
            for tag in $tags; do
                echo -n "\"$tag\": "
                get_for_tag "$tag"
                echo ","
            done
        ) | head -n -1
        echo "}"
    ) | jq .
else
    get_for_tag "$1"
fi


