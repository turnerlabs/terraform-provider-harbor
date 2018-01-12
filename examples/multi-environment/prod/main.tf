provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment_env" "app" {
  shipment    = "my-app"
  environment = "prod"
  barge       = "ent-prod"
  replicas    = 8 
  monitored   = true

  container {
    name = "my-app"

    port {
      protocol    = "http"
      public_port = 80
      value       = 5000
      healthcheck = "/health"
    }
  }
}

output "dns_name" {
  value = "${harbor_shipment_env.app.dns_name}"
}