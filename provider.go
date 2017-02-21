package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider returns a terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credential": {
				Type:        schema.TypeString,
				Optional:    false,
				Default:     "",
				Description: "credentials to manage harbor shipments",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			// "harbor_shipment": resourceHarborShipment(),
		},
		ConfigureFunc: providerConfigure,
	}
}


func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	creds := d.Get("access_key").(string)

	//todo: acquire valid harbor token

	return nil
}