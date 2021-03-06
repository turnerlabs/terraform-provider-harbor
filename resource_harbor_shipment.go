package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceHarborShipment() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborShipmentCreate,
		Read:   resourceHarborShipmentRead,
		Update: resourceHarborShipmentUpdate,
		Delete: resourceHarborShipmentDelete,
		Exists: resourceHarborShipmentExists,
		Importer: &schema.ResourceImporter{
			State: resourceHarborShipmentImport,
		},

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
	auth := meta.(*harborMeta).auth

	shipment := Shipment{
		Name:  d.Get("shipment").(string),
		Group: d.Get("group").(string),
	}

	//POST /v1/shipments
	writeMetric(metricShipmentCreate)
	uri := shipitURI("/v1/shipments")
	res, _, err := create(auth.Username, auth.Token, uri, shipment)
	if err != nil && len(err) > 0 {
		return err[0]
	}
	if res.StatusCode != http.StatusCreated {
		newErr := errors.New("unable to create shipment: " + err[0].Error())
		writeMetricError(metricShipmentCreate, newErr)
		check(newErr)
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
		writeMetricErrorString(metricShipmentCreate, "unable to create shipment envvar: "+err[0].Error())
		return err[0]
	}
	if res.StatusCode != http.StatusCreated {
		msg := "unable to create shipment envvar: status code = " + strconv.Itoa(res.StatusCode)
		writeMetricErrorString(metricShipmentCreate, msg)
		return errors.New(msg)
	}

	d.SetId(shipment.Name)

	return nil
}

func resourceHarborShipmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth
	writeMetric(metricShipmentDelete)
	uri := shipitURI("/v1/shipment/{shipment}", param("shipment", d.Id()))
	res, _, err := deleteHTTP(auth.Username, auth.Token, uri)
	if res.StatusCode != http.StatusOK {
		newErr := errors.New("shipment delete failed: status code = " + strconv.Itoa(res.StatusCode))
		writeMetricError(metricShipmentDelete, newErr)
		return newErr
	}
	if err != nil && len(err) > 0 {
		writeMetricErrorString(metricShipmentDelete, "shipment delete failed: "+err[0].Error())
		return err[0]
	}

	return nil
}

func resourceHarborShipmentUpdate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth

	if d.HasChange("group") {

		data := Shipment{
			Group: d.Get("group").(string),
		}

		//update the required shipment envvar for customer/group
		customerEnvVar := EnvVarPayload{
			Value: data.Group,
		}

		writeMetric(metricShipmentUpdate)
		uri := shipitURI("/v1/shipment/{shipment}/envVar/{envVar}",
			param("shipment", d.Id()),
			param("envVar", "CUSTOMER"))
		res, _, err := update(auth.Username, auth.Token, uri, customerEnvVar)
		if res.StatusCode != http.StatusOK {
			newErr := errors.New("shipment envvar update failed: status code = " + strconv.Itoa(res.StatusCode))
			writeMetricError(metricShipmentUpdate, newErr)
			return newErr
		}
		if err != nil && len(err) > 0 {
			return err[0]
		}

		//now update the shipment
		uri = shipitURI("/v1/shipment/{shipment}", param("shipment", d.Id()))
		res, _, err = update(auth.Username, auth.Token, uri, data)
		if res.StatusCode != http.StatusOK {
			newErr := errors.New("shipment update failed: status code = " + strconv.Itoa(res.StatusCode))
			writeMetricError(metricShipmentUpdate, newErr)
			return newErr
		}
		if err != nil && len(err) > 0 {
			newErr := errors.New("shipment update failed: " + err[0].Error())
			writeMetricError(metricShipmentUpdate, newErr)
			return newErr
		}
	}
	return nil
}

//has the resource been deleted outside of terraform?
func resourceHarborShipmentExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	auth := meta.(*harborMeta).auth
	shipment := GetShipment(auth.Username, auth.Token, d.Id())
	if shipment == nil {
		d.SetId("")
		return false, nil
	}
	return true, nil
}

//can assume resoure exists (since tf calls exists)
//remote data should be updated into the local data
func resourceHarborShipmentRead(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth
	shipment := GetShipment(auth.Username, auth.Token, d.Id())
	if shipment == nil {
		return errors.New("shipment doesn't exist")
	}

	d.Set("group", shipment.Group)

	return nil
}

func resourceHarborShipmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	//lookup and set the arguments
	auth := meta.(*harborMeta).auth
	writeMetric(metricShipmentImport)
	shipment := GetShipment(auth.Username, auth.Token, d.Id())
	if shipment == nil {
		newErr := errors.New("shipment doesn't exist")
		writeMetricError(metricShipmentImport, newErr)
		return nil, newErr
	}
	d.Set("shipment", shipment.Name)
	d.Set("group", shipment.Group)

	return []*schema.ResourceData{d}, nil
}
