package main

import (
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
				ForceNew: true,
			},
			"replicas": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

type environmentPayload struct {
	Name string `json:"name,omitempty"`
}

func resourceHarborShipmentEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	shipment := d.Get("shipment").(string)
	environment := d.Get("environment").(string)
	auth := meta.(Auth)

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
	d.SetId(fmt.Sprintf("shipment/%s/environment/%s", shipment, environment))
	return nil
}

func resourceHarborShipmentEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	// uri := shipmentURI(d.Id())
	// res, body, err := gorequest.New().Get(uri).EndBytes()
	// if err != nil {
	// 	return err[0]
	// }
	// if res.StatusCode != 200 {
	// 	return errors.New("get shipment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	// }

	// var result shipmentPayload
	// unmarshalErr := json.Unmarshal(body, &result)
	// if unmarshalErr != nil {
	// 	return unmarshalErr
	// }

	return nil
}

func resourceHarborShipmentEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	// auth := meta.(Auth)

	// //todo: cleanup -> set replicas=0/trigger

	// uri := shipmentURI(d.Id())
	// _, _, err := gorequest.New().Delete(uri).
	// 	Set("x-username", auth.Username).
	// 	Set("x-token", auth.Token).
	// 	End()
	// if err != nil {
	// 	return err[0]
	// }

	return nil
}

func resourceHarborShipmentEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {

	// if d.HasChange("group") {
	// 	_, newGroup := d.GetChange("group")

	// 	auth := meta.(Auth)

	// 	data := shipmentPayload{
	// 		Group: newGroup.(string),
	// 	}

	// 	uri := shipmentURI(d.Id())
	// 	res, _, err := gorequest.New().Put(uri).
	// 		Set("x-username", auth.Username).
	// 		Set("x-token", auth.Token).
	// 		Send(data).
	// 		End()

	// 	if err != nil {
	// 		return err[0]
	// 	}

	// 	if res.StatusCode != 200 {
	// 		return errors.New("update shipment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	// 	}
	// }

	return nil
}
