### Multiple Environments

This example shows how you can split up multiple environments.  The parent [`harbor_shipment`](main.tf) resource lives in the root (which is a good place to manage other project-wide resources, like terraform remote state, DNS zones, etc.) and the `harbor_shipment_env` [environments](dev/main.tf) live in sub directories.