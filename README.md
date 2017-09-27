terraform-provider-harbor
==========================

A [Terraform](https://www.terraform.io/) provider for managing [Harbor](https://github.com/turnerlabs/harbor) resources.


Benefits:

- infrastructure as code (versionable and reproducible infrastructure)
- all of your infrastructure declared in a single place, format, command
- native integration with the vast landscape of existing terraform providers
- outputs managed load balancer information for integration with route 53
- aws role integration
- tag integration
- works with changes made in GUI or CLI
- leverages terraform's first class support state synchronization


#### usage example

```terraform
provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "app" {
  shipment = "mss-poc-terraform"
  group    = "my-team"
}

resource "harbor_shipment_env" "prod" {
  shipment             = "${harbor_shipment.app.id}"
  environment          = "dev"
  barge                = "ent-prod"
  replicas             = 3
  monitored            = false
  healthcheck_timeout  = 1
  healthcheck_interval = 10

  container {
    name = "container"

    port {
      name         = "PORT"
      protocol     = "http"
      value        = 3000
      public_port  = 80
      health_check = "/health"
    }
  }  
  
  logShipping {
    type     = "logzio"
    endpoint = "http://xyz"
  }
}

data "aws_elb_hosted_zone_id" "region" {}

resource "aws_route53_record" "root" {
  zone_id = "${aws_route53_zone.my_app.zone_id}"
  name    = "my-app.turnerapps.com"
  type    = "A"

  alias {
    name                   = "${harbor_shipment_env.prod.lb_dns_name}"
    zone_id                = "${data.aws_elb_hosted_zone_id.region.id}"
    evaluate_target_health = false
  }
}
```
