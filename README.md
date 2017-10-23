terraform-provider-harbor
==========================

A [Terraform](https://www.terraform.io/) provider for managing [Harbor](https://github.com/turnerlabs/harbor) resources.

### Benefits

- infrastructure as code (versionable and reproducible infrastructure)
- all of your infrastructure declared in a single place, format, command
- native integration with the vast landscape of existing terraform providers
- user doesn't have to understand idiosyncrasies of ShipIt and Trigger, and what types of changes require setting replicas = 0 and triggering
- HTTPS/TLS setup via IaC
- AWS Tag integration
- AWS Route53 integration (outputs managed load balancer information)
- AWS ACM integration
- AWS Role integration (coming)
- works with changes made in GUI or CLI
- leverages terraform's first class support state synchronization

[![asciicast](https://asciinema.org/a/IxeoNYl7RmmSTwYMURtODytHO.png)](https://asciinema.org/a/IxeoNYl7RmmSTwYMURtODytHO?autoplay=1)

### Installation

- install plugin
```
curl -s get-tf.harbor.turnerlabs.io | sh
```

- install [harbor-compose](https://github.com/turnerlabs/harbor-compose) which is required for authentication and for deploying your application images and environment variables on to the infrastructure
```
curl -s get-cli.harbor.turnerlabs.io | sh
```


### Usage example

```terraform
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
}

output "dns_name" {
  value = "${harbor_shipment_env.dev.dns_name}"
}
```

### Other examples

- [DNS and SSL](examples/dns-ssl)