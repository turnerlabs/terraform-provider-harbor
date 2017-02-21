package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	harborauth "github.com/turnerlabs/harbor-auth-client"
)

var authURL = "https://auth.services.dmtio.net"

//Auth struc
type Auth struct {
	Version  string `json:"version"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// Provider returns a terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "credentials to manage harbor shipments",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"harbor_shipment_environment": resourceHarborShipmentEnvironment(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	//validate harbor token
	creds := d.Get("credentials").(string)
	log.Println(creds)
	if len(creds) == 0 {
		return nil, errors.New("missing credentials")
	}

	//deserialize credentials
	var auth Auth
	err := json.Unmarshal([]byte(creds), &auth)
	if err != nil {
		return nil, err
	}

	//validate that credentials are still valid
	client, err := harborauth.NewAuthClient(authURL)
	if err != nil {
		return nil, err
	}

	isAuth, err := client.IsAuthenticated(auth.Username, auth.Token)
	if err != nil {
		return nil, err
	}
	if !isAuth {
		return nil, errors.New("token is not valid")
	}

	return creds, nil
}
