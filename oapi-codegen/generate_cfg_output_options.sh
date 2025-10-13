#!/usr/bin/env bash

GOOD_TAGS="AUTH FIREWALL INTERFACE SYSTEM SERVICES"

echo "# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json"
echo "package: pfsense_rest_v2"
echo "output: ../internal/api/pfsense_rest_v2.gen.go"
echo "generate:"
echo "  models: true"
echo "  client: true"
echo "output-options:"
echo "  include-tags:"
for tag in $GOOD_TAGS; do
    echo "    - $tag"
done

# Now exclude all SERVICES operations that aren't for static mapping
before='["^/api/v2/services/dhcp_server$", "^/api/v2/services/dhcp_server/.*$" ] as $keep'
from_tag="SERVICES"

echo "  exclude-operation-ids:"
jq -r "$before"' |
    .paths | with_entries(select(
        .value | with_entries(select(
            .value.tags[]? | index("'$from_tag'")
        )) | length > 0
    )) as $filtered_paths |
    $filtered_paths | with_entries(select(
        .key as $key |
        any($keep[]; . as $re | $key | test($re)) | not
    ))
    | to_entries | map(.value | to_entries | map(.value.operationId)) | flatten
    | map("    - " + .) | join("\n")
    ' \
    $(dirname $0)/openapi.json
