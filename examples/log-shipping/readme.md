### Log Shipping

#### logz.io

```hcl
provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

resource "harbor_shipment" "app" {
  shipment = "my-app"
  group    = "mss"
}

resource "harbor_shipment_env" "dev" {
  shipment    = "${harbor_shipment.app.id}"
  environment = "dev"
  barge       = "digital-sandbox"
  replicas    = 4
  monitored   = false

  container {
    name = "my-app"

    port {
      protocol    = "http"
      public_port = 80
      value       = 5000
      healthcheck = "/health"
    }
  }

  # ship logs to logz.io
  log_shipping {
    provider = "logzio"
    endpoint = "https://listener.logz.io:8071?token=xxxxxx"
  }
}

output "dns_name" {
  value = "${harbor_shipment_env.dev.dns_name}"
}
```