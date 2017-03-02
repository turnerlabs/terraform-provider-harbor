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
	Name           string `json:"name,omitempty"`
	ParentShipment struct {
		Name string `json:"name,omitempty"`
	}
	Providers  []providerPayload  `json:"providers,omitempty"`
	Containers []containerPayload `json:"containers,omitempty"`
}

type environmentPayload struct {
	Name string `json:"name,omitempty"`
}

type providerPayload struct {
	Name     string `json:"name,omitempty"`
	Replicas int    `json:"replicas,omitempty"`
	Barge    string `json:"barge,omitempty"`
}

type containerPayload struct {
	Name    string          `json:"name,omitempty"`
	Image   string          `json:"image,omitempty"`
	Ports   []portPayload   `json:"ports,omitempty"`
	EnvVars []envVarPayload `json:"envVars,omitempty"`
}

type envVarPayload struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
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

	//POST /v1/shipment/:Shipment/environments
	uri := fullyQualifiedURI(shipment) + "/environments"
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

	//use the uri fragment as the id (shipment/foo/environment/dev)
	id := fmt.Sprintf("%s/environment/%s", shipment, environment)

	//now create related provider resource that maintains the barge and replicas
	payload := providerPayload{
		Name:     provider,
		Replicas: replicas,
		Barge:    barge,
	}

	//POST /v1/shipment/:Shipment/environment/:Environment/providers
	uri = fullyQualifiedURI(id + "/providers")
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

	uri := fullyQualifiedURI(d.Id())
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

	d.Set("environment", result.Name)

	prov := result.Providers[0]
	d.Set("barge", prov.Barge)
	d.Set("replicas", prov.Replicas)

	return nil
}

func resourceHarborShipmentEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	//todo: cleanup -> set replicas=0/trigger

	//maybe should delete provider?

	//DELETE /v1/shipment/:Shipment/environment/:name
	uri := fullyQualifiedURI(d.Id())
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

		payload := providerPayload{
			Replicas: replicas,
			Barge:    barge,
		}

		auth := meta.(Auth)

		//PUT /v1/shipment/:Shipment/environment/:Environment/provider/:name
		uri := fullyQualifiedURI(fmt.Sprintf("%s/provider/%s", d.Id(), provider))
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
