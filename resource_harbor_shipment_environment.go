package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/parnurzeal/gorequest"
)

func resourceHarborShipmentEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborShipmentEnvironmentCreate,
		Read:   resourceHarborShipmentEnvironmentRead,
		Update: resourceHarborShipmentEnvironmentUpdate,
		Delete: resourceHarborShipmentEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"shipment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"barge": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"replicas": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

var provider = "ec2"

type shipmentEnvironment struct {
	Name           string            `json:"name,omitempty"`
	Providers      []providerPayload `json:"providers,omitempty"`
	ParentShipment struct {
		Name string `json:"name,omitempty"`
	}
}

type environmentPayload struct {
	Name string `json:"name,omitempty"`
}

type providerPayload struct {
	Name     string `json:"name,omitempty"`
	Replicas int    `json:"replicas,omitempty"`
	Barge    string `json:"barge,omitempty"`
}

func resourceHarborShipmentEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	shipment := d.Get("shipment").(string)
	environment := d.Get("environment").(string)
	barge := d.Get("barge").(string)
	replicas := d.Get("replicas").(int)
	auth := meta.(Auth)

	//first create the environment resource
	data := environmentPayload{
		Name: environment,
	}

	uri := fmt.Sprintf("%s/v1/shipment/%s/environments", shipItURI, shipment)
	res, _, err := gorequest.New().Post(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create environment api returned " + strconv.Itoa(res.StatusCode))
	}

	id := fmt.Sprintf("shipment/%s/environment/%s", shipment, environment)

	//now create related provider resource that maintains the barge and replicas
	//POST /v1/shipment/:Shipment/environment/:Environment/providers
	payload := providerPayload{
		Name:     provider,
		Replicas: replicas,
		Barge:    barge,
	}

	uri = fmt.Sprintf("%s/v1/%s/providers", shipItURI, id)
	res, _, err = gorequest.New().Post(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(payload).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create provider api returned " + strconv.Itoa(res.StatusCode))
	}

	d.SetId(id)
	return nil
}

func resourceHarborShipmentEnvironmentRead(d *schema.ResourceData, meta interface{}) error {

	uri := fmt.Sprintf("%s/v1/%s", shipItURI, d.Id())
	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		return err[0]
	}
	if res.StatusCode != 200 {
		return errors.New("get environment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	}

	var result shipmentEnvironment
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	d.Set("shipment", result.ParentShipment.Name)
	d.Set("environment", result.ParentShipment)
	d.Set("barge", result.Providers[0].Barge)
	d.Set("replicas", result.Providers[0].Replicas)

	return nil
}

func resourceHarborShipmentEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	//todo: cleanup -> set replicas=0/trigger

	//maybe should delete provider?

	//DELETE /v1/shipment/:Shipment/environment/:name
	uri := fmt.Sprintf("%s/v1/%s", shipItURI, d.Id())
	_, _, err := gorequest.New().Delete(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		End()
	if err != nil {
		return err[0]
	}

	return nil
}

func resourceHarborShipmentEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {

	//changing barge or replicas requires a trigger
	if d.HasChange("barge") || d.HasChange("replicas") {

		_, newBarge := d.GetChange("barge")
		barge := newBarge.(string)
		_, newReplicas := d.GetChange("replicas")
		replicas := newReplicas.(int)

		//moving barges requires deleting the ELB
		if d.HasChange("barge") {
			//todo: cleanup -> set replicas=0/trigger
		}

		//PUT /v1/shipment/:Shipment/environment/:Environment/provider/:name
		payload := providerPayload{
			Replicas: replicas,
			Barge:    barge,
		}
		auth := meta.(Auth)

		uri := fmt.Sprintf("%s/v1/%s/provider/%s", shipItURI, d.Id(), provider)
		res, _, err := gorequest.New().Put(uri).
			Set("x-username", auth.Username).
			Set("x-token", auth.Token).
			Send(payload).
			End()

		if err != nil {
			return err[0]
		}

		if res.StatusCode != 200 {
			return errors.New("update provider api returned " + strconv.Itoa(res.StatusCode))
		}

		//todo: trigger

		d.Set("barge", barge)
		d.Set("replicas", replicas)
	}

	return nil
}
