### terraform-provider-harbor

A terraform provider for managing harbor shipment-related resources.


#### usage

```terraform
# data provider that reads
data "harbor_compose_token" {
  credentials = "${file("~/.harbor/credentials")}"
}

# provider level settings like where to get credentials
provider "harbor" {
  username = "${data.harbor_compose_token.username}"
  token    = "${data.harbor_compose_token.token}"
}

# shipment shell
resource "harbor_shipment" "mss-poc-terraform" {
  shipment = "mss-poc-terraform"
  group    = "mss"
}

# define environments
resource "harbor_shipment_environment" "dev" {
  shipment    = "${harbor_shipment.mss-poc-terraform.id}"
  environment = "dev"
  barge       = "digital-sandbox"
  replicas    = 2
}

# define environments
resource "harbor_shipment_environment" "qa" {
  shipment    = "${harbor_shipment.mss-poc-terraform.id}"
  environment = "qa"
  barge       = "digital-sandbox"
  replicas    = 2
}

# associate a container with the shipment
resource "harbor_container" "web" {
  environment = "${harbor_shipment_environment.dev.id}"
  name        = "web"
  image       = "quay.io/turner/web:1.0"
}

# associate a container with the shipment
resource "harbor_container" "worker" {
  environment = "${harbor_shipment_environment.dev.id}"
  name        = "worker"
  image       = "quay.io/turner/worker:1.0"
}

# envvars
resource "harbor_envvar" {
  container = "${harbor_container.web.id}"

  name  = "REDIS"
  value = "${elasticache_redis.redis.cache_nodes.0.address}"
}

resource "harbor_envvar" {
  container = "${harbor_container.web.id}"

  name  = "AWS_SECRET_KEY"
  value = "${aws_access_key.secret_key}"
  type  = "hidden"
}

# associate a port with a container
resource "harbor_port" "port" {
  container = "${harbor_container.web.id}"

  primary      = true
  protocol     = "http"
  value        = "8080"
  public_port  = "80"
  public_vip   = "false"
  health_check = "/hc"
}

# associate a port with a container
resource "harbor_port" "ssl" {
  container = "${harbor_container.web.id}"

  primary             = false
  protocol            = "https"
  value               = "443"
  public_port         = "80"
  public_vip          = "false"
  health_check        = "/hc"
  ssl_management_type = "acm"
  ssl_arn             = "${aws_acm_cert.app.arn}"
}

```