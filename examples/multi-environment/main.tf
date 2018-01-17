provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "app" {
  shipment = "my-app"
  group    = "mss"
}
