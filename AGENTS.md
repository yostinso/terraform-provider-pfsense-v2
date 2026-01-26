# AGENTS.md

## Project Overview
This repo contains a (work-in-progress) Terraform provider for PFSense networking devices.
It uses a library generated from the OpenAPI spec for the PFSenseV2 API but with a large chunk
of the provided services currently excluded to keep the library size reasonable.

The primary intent is to support configuring firewall rules (particularly port mapping) and DHCP
settings (static IP mappings, DHCP pools, hostname assignment).

It is a non-goal to support all PFSense features.

## Setup
This repo works within a devcontainer that providers terraform and various Golang debugging tools (like dlv) .
There are also tools for working with and generating the API library, including `jq` and `oapi-codegen`.
Sometimes you may need to run `go install .` to make sure dependencies are installed and the provider is
compiled and updated.

## Tools
### API library codegen
NB: This should only be run on user request.

The `oapi-codegen` folder contains the tooling for regenerating the OpenAPI library. All commands should
be run within the `oapi-codegen` folder.
1. If the user wants to get a fresh copy of the PFSensev2 spec, then run `./get_api_spec.sh`. This will fetch
   the spec, trim some XHR-enabling characters, and format the content through `jq`.
2. Once the spec is downloaded, the `cfg.yaml` used to configure `oapi-codegen` should be regenerated.
   The spec allows for tagging various API areas (i.e AUTH, FIREWALL, INTERFACE, etc.) but those are very
   broad strokes. `cfg.yml` allows include/exclude by tag and also by specific operation (i.e.
   getServicesACMEAccountKeyEndpoint).
   We want a few of the SERVICES operations, but we want to exclude most of them. Rather than
   manually maintaining this list, `generate_cfg_output_options.sh` looks at the spec and collects
   all the operations that _aren't_ related to DHCP and adds them to the exclusion list.

   Run `./generate_cfg_output_options.sh > cfg.yml` to generate a config.

   NEVER EDIT `cfg.yml` directly. Instead, make changes to the generator script.
3. Once the config is generated, you can now generated the library.
   Run `./generate.sh` and it will regenerate the generated library code in
   `/internal/api/pfsense_rest_v2.gen.go`.

   NEVER EDIT `pfsense_rest_v2.gen.go` directly. Instead, make changes to the config generator script.
   `pfsense_client_v2.go` is the Terraform wrapper for the library, and where any changes should be made
   to functionality.

## Running the provider
1. First, rebuild the provider if there are any changes with `go install .`
2. Next, the launch task "dlv - launch delve and capture TF_REATTACH_PROVIDERS" needs to be run.
3. Then you can launch terraform by sourcing the captured env variable and running the appropriate terraform
   command.
   For example, to run a `terraform plan`:
   ```
   . /workspaces/.vscode/.tf_provider_debug.env && terraform -chdir=examples/pfsensev2 plan
   ```

## Debugging
Debugging terraform providers requires a multi-step process:
1. Start up the provider API with `dlv debug` and capture the `TF_REATTACH_PROVIDERS` that is printed to stdout.
   - This is implemented in `tasks.json` in the "dlv - launch delve and capture TF_REATTACH_PROVIDERS" task.
2. Start terraform with the `TF_REATTACH_PROVIDERS` env var set to the value captured above.
   - This is implemented in the "dlv - start provider" task.
These two tasks have to be run in order; if you can't run them in the environment of vscode, just remind the
user that these need to be run in order via the task launcher.

## Running

## Expectations
* Changes should be incremental commits. Refer to `TODO.md` if the user is not providing the next task.
* Ask the developer if they want you to update `TODO.md` with new tasks when discussing future functionality.
* Tests should exist and should be run for every change.
