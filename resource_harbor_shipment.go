package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/parnurzeal/gorequest"
)

func resourceHarborShipment() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborShipmentCreate,
		Read:   resourceHarborShipmentRead,
		Update: resourceHarborShipmentUpdate,
		Delete: resourceHarborShipmentDelete,

		Schema: map[string]*schema.Schema{
			"shipment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"group": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

type shipmentPayload struct {
	Name  string `json:"name,omitempty"`
	Group string `json:"group,omitempty"`
}

func resourceHarborShipmentCreate(d *schema.ResourceData, meta interface{}) error {
	shipment := d.Get("shipment").(string)
	group := d.Get("group").(string)
	auth := meta.(Auth)

	data := shipmentPayload{
		Group: group,
		Name:  shipment,
	}

	res, _, err := gorequest.New().Post(shipItURI+"/v1/shipments/").
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create shipment api returned " + strconv.Itoa(res.StatusCode))
	}

	d.SetId(shipment)
	return nil
}

func shipmentURI(shipment string) string {
	return fmt.Sprintf("%s/v1/shipment/%s", shipItURI, shipment)
}

func resourceHarborShipmentRead(d *schema.ResourceData, meta interface{}) error {
	uri := shipmentURI(d.Id())
	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		return err[0]
	}
	if res.StatusCode != 200 {
		return errors.New("get shipment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	}

	var result shipmentPayload
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	d.Set("group", result.Group)

	return nil
}

func resourceHarborShipmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	//todo: cleanup -> set replicas=0/trigger

	uri := shipmentURI(d.Id())
	_, _, err := gorequest.New().Delete(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		End()
	if err != nil {
		return err[0]
	}

	return nil
}

func resourceHarborShipmentUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("group") {
		_, newGroup := d.GetChange("group")

		auth := meta.(Auth)

		data := shipmentPayload{
			Group: newGroup.(string),
		}

		uri := shipmentURI(d.Id())
		res, _, err := gorequest.New().Put(uri).
			Set("x-username", auth.Username).
			Set("x-token", auth.Token).
			Send(data).
			End()

		if err != nil {
			return err[0]
		}

		if res.StatusCode != 200 {
			return errors.New("update shipment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
		}
	}

	return nil
}
