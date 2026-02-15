# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Summary

Terraform provider for pfSense v2 networking devices, built with Go and the Terraform Plugin Framework. Manages firewall rules (port mapping) and DHCP settings (static IPs, pools, hostnames). See `AGENTS.md` for detailed project context and workflow instructions.

## Common Commands

```bash
# Build and install the provider
go install .

# Run tests
go test ./...

# Run linter
golangci-lint run

# Regenerate API library (run from oapi-codegen/ directory)
./get_api_spec.sh                      # Fetch fresh OpenAPI spec
./generate_cfg_output_options.sh > cfg.yaml  # Generate config
./generate.sh                          # Generate Go client code
```

## Architecture

```
main.go → provider.go (ScaffoldingProvider)
                ├── Resources/DataSources (internal/provider/*.go)
                └── PFSenseClientV2 (internal/api/pfsense_client_v2.go)
                        └── Generated API (pfsense_rest_v2.gen.go)
```

**Key directories:**
- `internal/provider/` - Terraform resources, data sources, provider config
- `internal/api/` - API client wrapper and generated code
- `oapi-codegen/` - API generation tooling
- `examples/pfsensev2/` - Working example Terraform config

## Critical Files

**Edit these:**
- `internal/provider/*.go` - Implement Terraform resources/data sources
- `internal/api/pfsense_client_v2.go` - API wrapper methods

**Never edit directly:**
- `internal/api/pfsense_rest_v2.gen.go` - Generated (10MB+), regenerate via scripts
- `oapi-codegen/cfg.yaml` - Generated, modify `generate_cfg_output_options.sh` instead

## Provider Authentication

Configured via Terraform or environment variables:
- URL: `PFSENSEV2_URL`
- Basic Auth: `PFSENSEV2_USERNAME` + `PFSENSEV2_PASSWORD`
- API Key Auth: `PFSENSEV2_TOKEN`
- Skip TLS verify: `PFSENSEV2_INSECURE`

## Debugging

Requires VS Code tasks (see `AGENTS.md` for details):
1. Run "dlv - launch delve and capture TF_REATTACH_PROVIDERS" task
2. Run terraform with captured env: `. /workspaces/.vscode/.tf_provider_debug.env && terraform -chdir=examples/pfsensev2 plan`
