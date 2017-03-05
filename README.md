### terraform-provider-harbor

A Terraform provider for managing Harbor resources.

This provider currently maps Terraform's apply/destory CRUD framework on to ShipIt's REST API.  Still need to look at integration with the Trigger API.

The provider currently adds value by:

- infrastructure as code (verionable and reproducible infrastructure)
- native integration with the vast landscape of existing terraform providers
- all of your infrastructure declared in a single place, format, command
- attach containers to your shipment as terraform modules


#### usage example

```terraform
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
  value     = "${module.elasticache_redis.endpoint}"
}

resource "harbor_envvar" "access-key" {
  container = "${harbor_container.web.id}"
  name      = "AWS_ACCESS_KEY"
  value     = "${module.s3-user.iam_access_key_id}"
  type      = "hidden"
}

resource "harbor_envvar" "secret-key" {
  container = "${harbor_container.web.id}"
  name      = "AWS_SECRET_KEY"
  value     = "${module.s3-user.iam_access_key_secret}"
  type      = "hidden"
}

```