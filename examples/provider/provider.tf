terraform {
  required_providers {
    pfsense-v2 = {
      source  = "registry.terraform.io/yostinso/pfsense-v2"
      # version = "0.1.0"
    }
  }
}
provider "pfsense-v2" {

}
