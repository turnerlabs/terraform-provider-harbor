provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}

data "harbor_loadbalancer" "main" {
  shipment    = "harbor-telemetry"
  environment = "dev"
}

output "dns_name" {
  value = "${data.harbor_loadbalancer.main.dns_name}"
}

output "name" {
  value = "${data.harbor_loadbalancer.main.name}"
}

output "type" {
  value = "${data.harbor_loadbalancer.main.type}"
}

output "arn" {
  value = "${data.harbor_loadbalancer.main.arn}"
}

output "hosted_zone_id" {
  value = "${data.harbor_loadbalancer.main.hosted_zone_id}"
}
