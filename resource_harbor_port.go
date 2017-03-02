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

func resourceHarborPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborPortCreate,
		Read:   resourceHarborPortRead,
		Update: resourceHarborPortUpdate,
		Delete: resourceHarborPortDelete,

		Schema: map[string]*schema.Schema{
			"container": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"primary": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "http",
			},
			"value": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"public_port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  80,
			},
			"public_vip": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"external": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"health_check": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"enable_proxy_protocol": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"private_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_key_certificate": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"certificate_chain": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_arn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_management_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

type portPayload struct {
	Name        string `json:"name,omitempty"`
	Value       int    `json:"value,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Healthcheck string `json:"healthcheck,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
	External    bool   `json:"external,omitempty"`
	PublicVip   bool   `json:"public_vip,omitempty"`
	PublicPort  int    `json:"public_port,omitempty"`
}

func resourceHarborPortCreate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

	//read user data
	container := d.Get("container").(string)
	primary := d.Get("primary").(bool)
	protocol := d.Get("protocol").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(int)
	external := d.Get("external").(bool)
	publicPort := d.Get("public_port").(int)
	publicVip := d.Get("public_vip").(bool)
	healthCheck := d.Get("health_check").(string)
	// enableProxyProtocol := d.Get("enable_proxy_protocol").(bool)
	// publicKeyCert := d.Get("public_key_certificate").(string)
	// certChain := d.Get("certificate_chain").(string)
	// sslArn := d.Get("ssl_arn").(string)
	// sslManagementType := d.Get("ssl_management_type").(string)

	//create the container resource
	data := portPayload{
		Name:        name,
		Value:       value,
		Protocol:    protocol,
		Healthcheck: healthCheck,
		Primary:     primary,
		External:    external,
		PublicVip:   publicVip,
		PublicPort:  publicPort,
	}

	//POST /v1/shipment/:Shipment/environment/:Environment/container/:Container/ports
	uri := fullyQualifiedURI(container) + "/ports"
	res, _, err := gorequest.New().Post(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create port api returned " + strconv.Itoa(res.StatusCode))
	}

	//use the uri fragment as the id (shipment/foo/environment/dev/container/bar/port/foo)
	d.SetId(fmt.Sprintf("%s/port/%s", container, name))

	return nil
}

func resourceHarborPortRead(d *schema.ResourceData, meta interface{}) error {

	//the id of this resource is the port uri (e.g., shipment/foo/environment/dev/container/bar/port/foo)
	//unfortunately, the server does not implement a get on this uri so we need to look for it
	//in the shipment/environment resource
	parts := strings.Split(d.Id(), "/")
	shipmentEnvURI := strings.Join(parts[:4], "/")
	containerName := parts[5]
	portName := parts[len(parts)-1]

	//fetch the port by id
	uri := fullyQualifiedURI(shipmentEnvURI)
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

	//now look for matching port by name
	var matchingPort *portPayload
	for _, port := range matchingContainer.Ports {
		if port.Name == portName {
			matchingPort = &port
			break
		}
	}

	if matchingPort == nil {
		return nil
	}

	//found it
	d.Set("primary", matchingPort.Primary)
	d.Set("protocol", matchingPort.Protocol)
	d.Set("name", matchingPort.Name)
	d.Set("value", matchingPort.Value)
	d.Set("external", matchingPort.External)
	d.Set("public_port", matchingPort.PublicPort)
	d.Set("public_vip", matchingPort.PublicVip)
	d.Set("health_check", matchingPort.Healthcheck)
	// enableProxyProtocol := d.Get("enable_proxy_protocol").(bool)
	// publicKeyCert := d.Get("public_key_certificate").(string)
	// certChain := d.Get("certificate_chain").(string)
	// sslArn := d.Get("ssl_arn").(string)
	// sslManagementType := d.Get("ssl_management_type").(string)

	return nil
}

func resourceHarborPortUpdate(d *schema.ResourceData, meta interface{}) error {

	//read user data
	primary := d.Get("primary").(bool)
	protocol := d.Get("protocol").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(int)
	external := d.Get("external").(bool)
	publicPort := d.Get("public_port").(int)
	publicVip := d.Get("public_vip").(bool)
	healthCheck := d.Get("health_check").(string)
	// enableProxyProtocol := d.Get("enable_proxy_protocol").(bool)
	// publicKeyCert := d.Get("public_key_certificate").(string)
	// certChain := d.Get("certificate_chain").(string)
	// sslArn := d.Get("ssl_arn").(string)
	// sslManagementType := d.Get("ssl_management_type").(string)

	data := portPayload{
		Name:        name,
		Value:       value,
		Protocol:    protocol,
		Healthcheck: healthCheck,
		Primary:     primary,
		External:    external,
		PublicVip:   publicVip,
		PublicPort:  publicPort,
	}

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
		return errors.New("update port api returned " + strconv.Itoa(res.StatusCode) + " for " + d.Id())
	}

	return nil
}

func resourceHarborPortDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(Auth)

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
		return errors.New("delete port api returned " + strconv.Itoa(res.StatusCode) + " for " + d.Id())
	}

	return nil
}
