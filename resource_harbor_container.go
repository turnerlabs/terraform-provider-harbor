package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/parnurzeal/gorequest"
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
	auth := meta.(Auth)

	environment := d.Get("environment").(string)
	name := d.Get("name").(string)
	image := d.Get("image").(string)

	//create the container resource
	data := containerPayload{
		Name:  name,
		Image: image,
	}

	//POST /v1/shipment/:Shipment/environment/:Environment/containers
	uri := fullyQualifiedURI(environment) + "/containers"
	res, _, err := gorequest.New().Post(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create container api returned " + strconv.Itoa(res.StatusCode))
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
	uri := fullyQualifiedURI(environment)
	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		return err[0]
	}
	if res.StatusCode == 404 {
		return nil
	} else if res.StatusCode != 200 {
		return errors.New("get environment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	}

	var result shipmentEnvironment
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	//try to find container in environment resource by container name
	if len(result.Containers) == 0 {
		return nil
	}

	var matchingContainer *containerPayload
	for _, container := range result.Containers {
		if container.Name == containerName {
			matchingContainer = &container
			break
		}
	}

	if matchingContainer == nil {
		return nil
	}

	//found it
	d.Set("image", matchingContainer.Image)

	return nil
}

func resourceHarborContainerUpdate(d *schema.ResourceData, meta interface{}) error {

	//image is the only updateable field
	if d.HasChange("image") {
		_, newImage := d.GetChange("image")

		data := containerPayload{
			Image: newImage.(string),
		}

		//PUT /v1/shipment/:Shipment/environment/:Environment/container/:name
		auth := meta.(Auth)
		uri := fullyQualifiedURI(d.Id())
		res, _, err := gorequest.New().Put(uri).
			Set("x-username", auth.Username).
			Set("x-token", auth.Token).
			Send(data).
			End()

		if err != nil {
			return err[0]
		}

		if res.StatusCode != 200 {
			return errors.New("update container api returned " + strconv.Itoa(res.StatusCode) + " for " + d.Id())
		}
	}

	return nil
}

func resourceHarborContainerDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	//DELETE /v1/shipment/:Shipment/environment/:Environment/container/:name
	uri := fullyQualifiedURI(d.Id())
	res, _, err := gorequest.New().Delete(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		End()
	if err != nil {
		return err[0]
	}

	if res.StatusCode == 404 || res.StatusCode == 422 {
		return nil
	} else if res.StatusCode != 200 {
		return errors.New("delete container api returned " + strconv.Itoa(res.StatusCode) + " for " + d.Id())
	}

	return nil
}
