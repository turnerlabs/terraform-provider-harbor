package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	harborauth "github.com/turnerlabs/harbor-auth-client"
)

var authURI = "https://auth.services.dmtio.net"
var shipItURI = "http://shipit.services.dmtio.net/v1"

func fullyQualifiedURI(id string) string {
	return fmt.Sprintf("%s/%s", shipItURI, id)
}

//Auth struct
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
				Description: "Harbor credentials. Run harbor-compose login to populate.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"harbor_shipment":             resourceHarborShipment(),
			"harbor_shipment_environment": resourceHarborShipmentEnvironment(),
			"harbor_container":            resourceHarborContainer(),
			"harbor_port":                 resourceHarborPort(),
			"harbor_envvar":               resourceHarborEnvvar(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"harbor_elb": dataSourceHarborElb(),
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
	client, err := harborauth.NewAuthClient(authURI)
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

	return &auth, nil
}
