provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

provider "aws" {
  region = "us-east-1"
}

module "s3-user" {
  source = "github.com/turnerlabs/terraform-s3-user?ref=v1.2"

  bucket_name = "terraform-harbor-test"
  user_name   = "srv_dev_terraform-harbor-test"

  tag_team          = "my-team"
  tag_contact-email = "my-team@turner.com"
  tag_application   = "my-app"
  tag_environment   = "dev"
  tag_customer      = "my-customer"
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
