# TODO

* [x] Try to get an output from reading current state; see https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-data-source-read
    * Attempted as part of `terraform -chdir=examples/pfsensev2 plan`
    * Provider no longer crashes during data source read.
    * Current environment result: successful `plan` output includes current state (e.g. `hostname = "ng4100"` and firewall rule list).

* [x] Significantly improve line and branch coverage for the existing current-state read methods
    * Focus on `GetBaseConfig`, `GetFirewallRules`, and data source `Read` success/error paths
    * Add targeted tests for nil responses, non-200 payload handling, and mapping behavior

* [x] Attempt to create a new firewall rule and then drop it
    * Implemented `CreateFirewallRule`, `GetFirewallRule`, `DeleteFirewallRule` in the API client
    * Implemented `PFSenseFirewallRuleResource` (Create/Read/Delete)
    * Added unit tests, HTTP happy/unhappy path tests, and an acceptance test for the create+destroy lifecycle

* Implement in-place rule updates via `PATCH /api/v2/firewall/rule`
    * Add `UpdateFirewallRule` to the API client with unit and HTTP tests
    * Wire up `Update` on `PFSenseFirewallRuleResource` (currently returns an unsupported error)

* Add an E2E acceptance test for the full rule lifecycle: create, update, delete
    * Depends on the update implementation above
    * Multi-step acceptance test: step 1 creates, step 2 modifies a field (e.g. description), step 3 removes the resource to trigger destroy

