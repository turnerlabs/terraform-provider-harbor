provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "mss-poc-terraform" {
  shipment = "mss-poc-terraform"
  group    = "mss"
}

resource "harbor_shipment_environment" "dev" {
  shipment    = "${harbor_shipment.mss-poc-terraform.id}"
  environment = "dev"
  barge       = "digital-sandbox"
  replicas    = 2
}
