provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment_environment" "mss-poc-terraform" {
  environment = "dev"
}
