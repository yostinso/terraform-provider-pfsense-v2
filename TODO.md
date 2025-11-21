# TODO

* Try to get an output from reading current state; see https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-data-source-read


Debugging:
go install .
Run debug from launch.json
export the TF_REATTACH_PROVIDERS line
run :
TF_LOG=DEBUG terraform -chdir=examples/pfsensev2 plan

