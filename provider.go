package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	harborauth "github.com/turnerlabs/harbor-auth-client"
)

var authURI = "https://auth.services.dmtio.net"
var shipItURI = "http://shipit.services.dmtio.net/v1"
var infrastructureURI = "http://localhost:8080/api"

func fullyQualifiedURI(id string) string {
	return fmt.Sprintf("%s/%s", shipItURI, id)
}

func fullyQualifiedInfrastructureURI(id string) string {
	return fmt.Sprintf("%s/%s", infrastructureURI, id)
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
			"harbor_shipment":     resourceHarborShipment(),
			"harbor_shipment_env": resourceHarborShipmentEnv(),
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

	success, err := client.IsAuthenticated(auth.Username, auth.Token)
	if err != nil {
		if strings.Contains(err.Error(), "401 Unauthorized") {
			return nil, errors.New("Token has expired. Please run harbor-compose login")
		}
		return nil, err
	}
	if !success {
		return nil, errors.New("auth failed")
	}

	return &auth, nil
}
