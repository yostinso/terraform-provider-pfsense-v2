terraform {
  required_providers {
    pfsense-v2 = {
      source = "registry.terraform.io/yostinso/pfsense-v2"
      # version = "0.1.0"
    }
  }
}
provider "pfsense-v2" {
  url                 = "https://192.168.2.1"
  insecure            = true
  api_client_token    = "c1436e040ccc971c179ef343d58577f3"
}

data "pfsense-v2_configs" "some-output" {}

output "system_info" {
  value = data.pfsense-v2_configs.some-output
}
