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

	data := shipmentPayload{
		Name:  d.Get("shipment").(string),
		Group: d.Get("group").(string),
	}

	//POST /v1/shipments
	err := create("shipments", meta.(*Auth), data)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("shipment/%s", data.Name)

	//create the required shipment envvar for customer/group
	customerEnvVar := envVarPayload{
		Name:  "CUSTOMER",
		Value: data.Group,
	}

	//POST /v1/shipment/:Shipment/envVars
	err = create(id+"/envVars", meta.(*Auth), customerEnvVar)
	if err != nil {
		return err
	}

	//use the uri fragment as the id (shipment/foo)
	d.SetId(id)

	return nil
}

func resourceHarborShipmentRead(d *schema.ResourceData, meta interface{}) error {

	uri := fullyQualifiedURI(d.Id())
	res, body, err := gorequest.New().Get(uri).EndBytes()
	if err != nil {
		return err[0]
	}
	if res.StatusCode == 404 {
		return nil
	} else if res.StatusCode != 200 {
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
	//todo: cleanup -> set replicas=0/trigger
	//customer envvar should get cascade deleted
	return delete(d.Id(), meta.(*Auth))
}

func resourceHarborShipmentUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("group") {

		data := shipmentPayload{
			Group: d.Get("group").(string),
		}

		//update the required shipment envvar for customer/group
		customerEnvVar := envVarPayload{
			Value: data.Group,
		}

		uri := d.Id() + "/envVar/CUSTOMER"
		err := update(uri, meta.(*Auth), customerEnvVar)
		if err != nil {
			return err
		}

		//update the shipment
		return update(d.Id(), meta.(*Auth), data)
	}
	return nil
}
