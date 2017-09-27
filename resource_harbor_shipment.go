package main

import (
	"errors"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
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

func resourceHarborShipmentCreate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*Auth)

	shipment := Shipment{
		Name:  d.Get("shipment").(string),
		Group: d.Get("group").(string),
	}

	//POST /v1/shipments
	uri := shipitURI("/v1/shipments")
	res, _, err := create(auth.Username, auth.Token, uri, shipment)
	if err != nil && len(err) > 0 {
		return err[0]
	}
	if res.StatusCode != http.StatusCreated {
		check(errors.New("unable to create shipment"))
	}

	//create the required shipment envvar for customer/group
	customerEnvVar := EnvVarPayload{
		Name:  "CUSTOMER",
		Value: shipment.Group,
	}

	//POST /v1/shipment/:Shipment/envVars
	uri = shipitURI("/v1/shipment/{shipment}/envVars", param("shipment", shipment.Name))
	res, _, err = create(auth.Username, auth.Token, uri, customerEnvVar)
	if err != nil && len(err) > 0 {
		return err[0]
	}
	if res.StatusCode != http.StatusCreated {
		return errors.New("unable to create shipment")
	}

	d.SetId(shipment.Name)

	return nil
}

func resourceHarborShipmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*Auth)
	uri := shipitURI("/v1/shipment/{shipment}", param("shipment", d.Id()))
	res, _, err := deleteHTTP(auth.Username, auth.Token, uri)
	if res.StatusCode != http.StatusOK {
		return errors.New("shipment delete failed")
	}
	if err != nil && len(err) > 0 {
		return err[0]
	}
	return nil
}

func resourceHarborShipmentUpdate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*Auth)

	if d.HasChange("group") {

		data := Shipment{
			Group: d.Get("group").(string),
		}

		//update the required shipment envvar for customer/group
		customerEnvVar := EnvVarPayload{
			Value: data.Group,
		}

		uri := shipitURI("/v1/shipment/{shipment}/envVar/{envVar}",
			param("shipment", d.Id()),
			param("envVar", "CUSTOMER"))
		res, _, err := update(auth.Username, auth.Token, uri, customerEnvVar)
		if res.StatusCode != http.StatusOK {
			return errors.New("shipment envvar update failed")
		}
		if err != nil && len(err) > 0 {
			return err[0]
		}

		//now update the shipment
		uri = shipitURI("/v1/shipment/{shipment}", param("shipment", d.Id()))
		res, _, err = update(auth.Username, auth.Token, uri, data)
		if res.StatusCode != http.StatusOK {
			return errors.New("shipment update failed")
		}
		if err != nil && len(err) > 0 {
			return err[0]
		}
	}
	return nil
}

func resourceHarborShipmentRead(d *schema.ResourceData, meta interface{}) error {
	if d.Id() == "" {
		return nil
	}

	auth := meta.(*Auth)
	shipment := GetShipment(auth.Username, auth.Token, d.Id())
	if shipment == nil {
		return nil
	}

	d.Set("group", shipment.Group)

	return nil
}
