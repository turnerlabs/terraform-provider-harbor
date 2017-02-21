# provider level settings like where to get credentials
provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

# define a shipment/environment
resource "harbor_shipment" "mss-poc-thingproxy" {
  group       = "mss"
  environment = "dev"
  barge       = "digital-sandbox"
  replicas    = "2"

  customer = ""
  property = ""
  project  = ""
  Product  = ""
}

# associate a container with the shipment
resource "harbor_container" "web" {
  shipment = "${harbor_shipment.mss-poc-thingproxy.id}"

  image = "registry.services.dmtio.net/kong:0.9.3"

  # environment_variables {
  #   REDIS         = "${elasticache_redis.redis.cache_nodes.0.address}"
  #   SHIP_LOGS     = "aws-elasticsearch"
  #   LOGS_ENDPOINT = "${aws_elasticsearch_domain.es.endpoint}"
  # }
}

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
