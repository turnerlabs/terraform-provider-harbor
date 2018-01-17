provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "app" {
  shipment = "iam-cert-example"
  group    = "mss"
}

resource "harbor_shipment_env" "dev" {
  shipment    = "${harbor_shipment.app.id}"
  environment = "dev"
  barge       = "digital-sandbox"
  replicas    = 4
  monitored   = true

  container {
    name = "iam-cert-example"

    port {
      protocol               = "https"
      public_port            = 443 
      value                  = 5000
      healthcheck            = "/health"
      ssl_management_type    = "iam"
      private_key            = "${file("./key.pem")}"
      public_key_certificate = "${file("./cert.pem")}"
      certificate_chain      = ""
    }
  }
}

output "dns_name" {
  value = "${harbor_shipment_env.dev.dns_name}"
}
