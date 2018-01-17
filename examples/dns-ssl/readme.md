### DNS and SSL example using AWS

- attaches a friendly DNS name to the shipment's default DNS
- configures TLS/SSL on the shipment by attaching an ACM certificate
- assumes an existing Route53 zone and issued ACM certificate
- ACM cert must exist in the same AWS account as the Harbor barge