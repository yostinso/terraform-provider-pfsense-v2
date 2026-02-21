# TODO

* [x] Try to get an output from reading current state; see https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-data-source-read
    * Attempted as part of `terraform -chdir=examples/pfsensev2 plan`
    * Provider no longer crashes during data source read.
    * Current environment result: successful `plan` output includes current state (e.g. `hostname = "ng4100"` and firewall rule list).

* [x] Significantly improve line and branch coverage for the existing current-state read methods
    * Focus on `GetBaseConfig`, `GetFirewallRules`, and data source `Read` success/error paths
    * Add targeted tests for nil responses, non-200 payload handling, and mapping behavior

* Attempt to create a new firewall rule and then drop it
    * Validate full lifecycle via Terraform (`apply` to create, then remove from config and `apply` to delete)
    * Add Terraform apply tests for firewall rule state reads/updates (rule present on first apply, absent on second apply)

