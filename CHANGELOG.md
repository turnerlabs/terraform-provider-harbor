## 0.6.0 (2018-02-07)

Features:

- Adds aws role support via `iam_role` ([#10](https://github.com/turnerlabs/terraform-provider-harbor/issues/16))

- Adds user/version to telemetry payload ([#41](https://github.com/turnerlabs/terraform-provider-harbor/issues/41))


## 0.5.0 (2018-01-17)

Features:

- Implement lbtype (support for ELB) ([#13](https://github.com/turnerlabs/terraform-provider-harbor/issues/13))

- output `build token` ([#16](https://github.com/turnerlabs/terraform-provider-harbor/issues/16))

- Add `harbor_loadbalancer` datasource that uses new trigger api rather than aws calls ([#12](https://github.com/turnerlabs/terraform-provider-harbor/issues/12))

- Support for IAM certs ([#11](https://github.com/turnerlabs/terraform-provider-harbor/issues/11))

- Adds more examples ([#40](https://github.com/turnerlabs/terraform-provider-harbor/issues/40))


Bug Fixes:

- Error running plan with output after importing ([#31](https://github.com/turnerlabs/terraform-provider-harbor/issues/31))

- Preserve build token on changes that involve downtime ([#33](https://github.com/turnerlabs/terraform-provider-harbor/issues/33))



## 0.4.0 (2017-11-10)

Features:

- Move healthcheck, timeout, interval to container level ([#5](https://github.com/turnerlabs/terraform-provider-harbor/issues/5))

- Support for multiple containers ([#7](https://github.com/turnerlabs/terraform-provider-harbor/issues/7))

- Support for log shipping ([#6](https://github.com/turnerlabs/terraform-provider-harbor/issues/6))

- Support for terraform import ([#9](https://github.com/turnerlabs/terraform-provider-harbor/issues/9))

- Support for annotations ([#14](https://github.com/turnerlabs/terraform-provider-harbor/issues/14))

- telemetry integration ([#20](https://github.com/turnerlabs/terraform-provider-harbor/issues/20))

- documentation ([#23](https://github.com/turnerlabs/terraform-provider-harbor/issues/23))



## 0.3.0 (2017-10-02)

Features:

- MVP (minimum viable product) of fully managed harbor infrastructure using ShipIt and Trigger APIs.


## 0.2.1 (2017-08-30)

Features:

- Improved error handling


## 0.2.0 (2017-08-28)

Features:

- Implemented `harbor_elb` datasource


## 0.1.0 (2017-03-13)

Features:

- Initial prototype that works with ShipIt API.