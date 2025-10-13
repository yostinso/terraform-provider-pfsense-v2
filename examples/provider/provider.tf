terraform {
  required_providers {
    pfsense-v2 = {
      source = "registry.terraform.io/yostinso/pfsense-v2"
      # version = "0.1.0"
    }
  }
}
provider "pfsense-v2" {
  url                 = "https://192.168.1.1"
  insecure            = true
  api_client_username = "admin"
  api_client_token    = "1234ABCD"
}
