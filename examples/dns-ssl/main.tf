provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

variable "barge" {
  default = "corp-sandbox"
}

variable "app" {
  default = "dns-ssl-example"
}

variable "friendly_dns" {
  default = "myapp.com"
}

provider "aws" {
  region  = "us-east-1"
  profile = "${var.barge}"
}

resource "harbor_shipment" "app" {
  shipment = "${var.app}"
  group    = "mss"
}

resource "harbor_shipment_env" "dev" {
  shipment             = "${harbor_shipment.app.id}"
  environment          = "dev"
  barge                = "${var.barge}"
  replicas             = 4
  monitored            = false

  container {
    name                 = "${var.app}"
    healthcheck          = "/health"
    healthcheck_timeout  = 1
    healthcheck_interval = 10

    port {
      primary             = "true"
      protocol            = "https"
      public_port         = 443
      value               = 5000
      healthcheck         = "/health"
      ssl_management_type = "acm"
      ssl_arn             = "${data.aws_acm_certificate.app.arn}"
    }
  }
}

data "aws_route53_zone" "app" {
  name = "${var.friendly_dns}."
}

resource "aws_route53_record" "dev" {
  zone_id = "${data.aws_route53_zone.app.zone_id}"
  type    = "CNAME"
  name    = "dev.${var.friendly_dns}"
  records = ["${harbor_shipment_env.dev.lb_dns_name}"]
  ttl     = "30"
}

data "aws_acm_certificate" "app" {
  domain   = "${var.friendly_dns}"
  statuses = ["ISSUED"]
}
