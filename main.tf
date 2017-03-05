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

resource "harbor_container" "web" {
  environment = "${harbor_shipment_environment.dev.id}"
  name        = "web"
  image       = "registry.services.dmtio.net/mss-poc-thingproxy:0.0.13-rc.42"
}

resource "harbor_port" "port" {
  container    = "${harbor_container.web.id}"
  name         = "80"
  protocol     = "http"
  value        = 3000
  public_port  = 80
  health_check = "/health"
}

resource "harbor_envvar" "REDIS" {
  container = "${harbor_container.web.id}"
  name      = "REDIS"
  value     = "foo"
}
