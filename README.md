terraform-provider-harbor
==========================

A [Terraform](https://www.terraform.io/) provider for managing [Harbor](https://github.com/turnerlabs/harbor) resources.

This provider currently maps Terraform's apply/destory CRUD framework on to ShipIt's REST API.  

Benefits:

- infrastructure as code (verionable and reproducible infrastructure)
- all of your infrastructure declared in a single place, format, command
- native integration with the vast landscape of existing terraform providers
- integrate reusable containers with your shipment as terraform modules


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

resource "harbor_port" "ssl" {
  container           = "${harbor_container.web.id}"
  name                = "ssl"
  protocol            = "https"
  public_port         = 443
  value               = 3000  
  health_check        = "/health"
  ssl_management_type = "acm"
  ssl_arn             = "${aws_acm_certificate.my-app.arn}"
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

### todo

- consider if/how to integrate trigger/deployments
- output endpoints and load balancer info