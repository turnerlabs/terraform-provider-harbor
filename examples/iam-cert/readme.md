### TLS/HTTPS example using an IAM certificate

- configures TLS/HTTPS on a shipment by attaching an IAM certificate

Obtain a certificate (example uses a self-signed certificate)

```
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365
```

Provision using terraform (be sure to use secure terraform remote state since the certificate private key will be stored there)

```
ssl_management_type    = "iam"
private_key            = "${file("./key.pem")}"
public_key_certificate = "${file("./cert.pem")}"
certificate_chain      = ""
```

```
terraform apply
```

For additional information on requesting a turner.com certificate and configuring the `certificate_chain`, please refer to [this post](http://blog.harbor.inturner.io/articles/shipment-ssl-howto/).