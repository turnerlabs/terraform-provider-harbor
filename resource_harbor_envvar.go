package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceHarborEnvvar() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborEnvvarCreate,
		Read:   resourceHarborEnvvarRead,
		Update: resourceHarborEnvvarUpdate,
		Delete: resourceHarborEnvvarDelete,

		Schema: map[string]*schema.Schema{
			"container": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "basic",
			},
		},
	}
}

func resourceHarborEnvvarCreate(d *schema.ResourceData, meta interface{}) error {

	//read user data
	container := d.Get("container").(string)

	//create the envvar resource
	data := envVarPayload{
		Name:  d.Get("name").(string),
		Value: d.Get("value").(string),
		Type:  d.Get("type").(string),
	}

	//POST /v1/shipment/:Shipment/environment/:Environment/container/:Container/envVars
	uri := fullyQualifiedURI(container) + "/envvars"
	err := create(uri, meta.(Auth), data)
	if err != nil {
		return err
	}

	//use the uri fragment as the id (shipment/foo/environment/dev/container/bar/envvar/foo)
	d.SetId(fmt.Sprintf("%s/envvar/%s", container, data.Name))

	return nil
}

func resourceHarborEnvvarRead(d *schema.ResourceData, meta interface{}) error {

	//the id of this resource is the port uri (e.g., shipment/foo/environment/dev/container/bar/envvar/foo)
	//unfortunately, the server does not implement a get on this uri so we need to look for it
	//in the shipment/environment resource
	shipmentEnvURI, containerName, envvarName := parseContainerResourceURI(d.Id())
	matchingContainer, err := readContainer(shipmentEnvURI, containerName, meta.(Auth))
	if err != nil {
		return err
	}

	//now look for matching envvar by name
	var matchingEnvVar *envVarPayload
	for _, envvar := range matchingContainer.EnvVars {
		if envvar.Name == envvarName {
			matchingEnvVar = &envvar
			break
		}
	}

	if matchingEnvVar == nil {
		return nil
	}

	//found it
	d.Set("value", matchingEnvVar.Value)
	d.Set("type", matchingEnvVar.Type)

	return nil
}

func resourceHarborEnvvarUpdate(d *schema.ResourceData, meta interface{}) error {

	//read user data
	data := envVarPayload{
		Value: d.Get("value").(string),
		Type:  d.Get("type").(string),
	}

	return update(d.Id(), meta.(Auth), data)
}

func resourceHarborEnvvarDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d.Id(), meta.(Auth))
}
