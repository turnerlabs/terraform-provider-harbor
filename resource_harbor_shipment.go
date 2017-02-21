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
			"barge": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type shipmentPayload struct {
	Name  string `json:"name,omitempty"`
	Group string `json:"group,omitempty"`
}

var shipItURI = "http://shipit.services.dmtio.net"

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

func resourceHarborShipmentRead(d *schema.ResourceData, meta interface{}) error {
	uri := fmt.Sprintf("%s/v1/shipment/%s", shipItURI, d.Id())
	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		return err[0]
	}
	if res.StatusCode != 200 {
		return errors.New("get shipment api returned " + strconv.Itoa(res.StatusCode))
	}

	var result shipmentPayload
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	return nil
}

func resourceHarborShipmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	uri := fmt.Sprintf("%s/v1/shipment/%s", shipItURI, d.Id())
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
	return nil
}
