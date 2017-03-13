package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceHarborContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborContainerCreate,
		Read:   resourceHarborContainerRead,
		Update: resourceHarborContainerUpdate,
		Delete: resourceHarborContainerDelete,

		Schema: map[string]*schema.Schema{
			"environment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHarborContainerCreate(d *schema.ResourceData, meta interface{}) error {

	environment := d.Get("environment").(string)
	name := d.Get("name").(string)
	image := d.Get("image").(string)

	//create the container resource
	data := containerPayload{
		Name:  name,
		Image: image,
	}

	//POST /v1/shipment/:Shipment/environment/:Environment/containers
	uri := environment + "/containers"
	err := create(uri, meta.(*Auth), data)
	if err != nil {
		return err
	}

	//use the uri fragment as the id (shipment/foo/environment/dev/container/bar)
	d.SetId(fmt.Sprintf("%s/container/%s", environment, name))

	return nil
}

func resourceHarborContainerRead(d *schema.ResourceData, meta interface{}) error {
	//the id of this resource is the container uri (e.g., shipment/foo/environment/dev/container/bar)
	//unfortunately, the server does not implement a get on this uri so we need to look for it
	//in the shipment/environment resource
	parts := strings.Split(d.Id(), "/")
	containerName := parts[len(parts)-1]

	//resource field is required and immutable since a container resource can't exist outside of an environment
	environment := d.Get("environment").(string)

	matchingContainer, err := readContainer(environment, containerName, meta.(*Auth))
	if err != nil {
		return err
	}

	//found it
	d.Set("image", matchingContainer.Image)

	return nil
}

func resourceHarborContainerUpdate(d *schema.ResourceData, meta interface{}) error {

	//image is the only updateable field
	if d.HasChange("image") {

		data := containerPayload{
			Image: d.Get("image").(string),
		}

		return update(d.Id(), meta.(*Auth), data)
	}
	return nil
}

func resourceHarborContainerDelete(d *schema.ResourceData, meta interface{}) error {
	return delete(d.Id(), meta.(*Auth))
}
