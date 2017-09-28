package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

const defaultBackendImageName = "quay.io/turner/turner-defaultbackend"
const defaultBackendImageVersion = "0.1.2"

func resourceHarborShipmentEnv() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborShipmentEnvironmentCreate,
		Read:   resourceHarborShipmentEnvironmentRead,
		Update: resourceHarborShipmentEnvironmentUpdate,
		Delete: resourceHarborShipmentEnvironmentDelete,
		Exists: resourceHarborShipmentEnvironmentExists,

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
			"monitored": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"healthcheck_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"healthcheck_interval": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"container": &schema.Schema{
				Description: "The list of containers for this shipment environment",
				Optional:    true,
				ForceNew:    true,
				MinItems:    1,
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"port": {
							Optional: true,
							ForceNew: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": &schema.Schema{
										Type:     schema.TypeBool,
										Required: true,
										ForceNew: true,
									},
									"protocol": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										Default:  "http",
										ForceNew: true,
									},
									"value": &schema.Schema{
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"public_port": &schema.Schema{
										Type:     schema.TypeInt,
										Optional: true,
										Default:  80,
										ForceNew: true,
									},
									"public": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
										ForceNew: true,
									},
									"external": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"healthcheck": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
										ForceNew: true,
									},
									"enable_proxy_protocol": &schema.Schema{
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
										ForceNew: true,
									},
									"ssl_arn": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"ssl_management_type": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										Default:  "iam",
										ForceNew: true,
									},
									"private_key": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"public_key_certificate": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"certificate_chain": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			//attributes
			"lb_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"lb_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"lb_arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"lb_dns_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"lb_hosted_zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHarborShipmentEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth
	shipmentName := d.Get("shipment").(string)
	environment := d.Get("environment").(string)

	//lookup the shipment in order to get the group/envvars (required for bulk creating env)
	shipment := GetShipment(auth.Username, auth.Token, shipmentName)
	if shipment == nil {
		return errors.New("shipment not found")
	}

	//is there an existing image to apply?
	existingImage := ""
	if meta.(*harborMeta).state["image"] != nil {
		existingImage = meta.(*harborMeta).state["image"].(string)
		if Verbose {
			log.Println("existing image = " + existingImage)
		}
	}

	//transform tf resource data into shipit model
	shipmentEnv, err := transformTerraformToShipmentEnvironment(d, existingImage, shipment.Group, shipment.EnvVars)
	if err != nil {
		return err
	}

	//add auth
	shipmentEnv.Username = auth.Username
	shipmentEnv.Token = auth.Token

	//debug print json
	if Verbose {
		b, _ := json.MarshalIndent(shipmentEnv, "\t", "\t")
		log.Println(string(b))
	}
	//return errors.New("debug")

	//post new shipment/environment
	SaveNewShipmentEnvironment(auth.Username, auth.Token, *shipmentEnv)

	//trigger shipment
	success, messages := Trigger(shipmentName, environment)
	if !success {
		failureMessage := ""
		for _, m := range messages {
			failureMessage += m + "\n"
		}
		return fmt.Errorf("trigger failed: %v", failureMessage)
	}

	//poll lb endpoint until it's ready
	var lbStatus *getLoadBalancerStatusResponse
	for {
		result, err := getLoadBalancerStatus(shipmentName, environment)
		if err != nil {
			return err
		}
		if result != nil {
			lbStatus = result
			break
		}
		//wait a few seconds
		time.Sleep(10 * time.Second)
	}
	if len(lbStatus.LoadBalancers) < 1 {
		return errors.New("no load balancers")
	}

	//output id
	d.SetId(fmt.Sprintf("%s::%s", shipmentEnv.ParentShipment.Name, shipmentEnv.Name))

	//output attributes
	d.Set("lb_name", lbStatus.LoadBalancers[0].LoadBalancerName)
	d.Set("lb_type", lbStatus.LoadBalancers[0].Type)
	d.Set("lb_arn", lbStatus.LoadBalancers[0].LoadBalancerArn)
	d.Set("lb_dns_name", lbStatus.LoadBalancers[0].DNSName)
	d.Set("lb_hosted_zone_id", lbStatus.LoadBalancers[0].CanonicalHostedZoneID)

	return nil
}

func idParts(id string) (string, string) {
	parts := strings.Split(id, "::")
	return parts[0], parts[1]
}

func resourceHarborShipmentEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	harborMeta := meta.(*harborMeta)
	auth := harborMeta.auth
	shipment, env := idParts(d.Id())

	shipmentEnv := GetShipmentEnvironment(auth.Username, auth.Token, shipment, env)
	if shipmentEnv == nil {
		return errors.New("shipment/environment doesn't exist")
	}

	//set replicas to 0 and trigger
	provider := ProviderPayload{
		Name:     providerEc2,
		Replicas: 0,
	}
	UpdateProvider(auth.Username, auth.Token, shipment, env, provider)

	//trigger shipment
	Trigger(shipment, env)

	//now delete from shipit
	DeleteShipmentEnvironment(auth.Username, auth.Token, shipment, env)

	//if deleting a shipment/env with an image other than the default backend,
	//then save the image to apply to creates in this same transaction
	if !strings.HasPrefix(shipmentEnv.Containers[0].Image, defaultBackendImageName) {
		harborMeta.state["image"] = shipmentEnv.Containers[0].Image
		if Verbose {
			log.Println("setting image = " + shipmentEnv.Containers[0].Image)
		}
	}

	return nil
}

//has the resource been deleted outside of terraform?
func resourceHarborShipmentEnvironmentExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	auth := meta.(*harborMeta).auth
	shipment, env := idParts(d.Id())
	shipmentEnv := GetShipmentEnvironment(auth.Username, auth.Token, shipment, env)
	if shipmentEnv == nil {
		d.SetId("")
		return false, nil
	}
	return true, nil
}

//can assume resoure exists (since tf calls exists)
//remote data should be updated into the local data
func resourceHarborShipmentEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth
	shipment, env := idParts(d.Id())
	shipmentEnv := GetShipmentEnvironment(auth.Username, auth.Token, shipment, env)
	if shipmentEnv == nil {
		return errors.New("shipment/environment doesn't exist")
	}

	//transform shipit model back to terraform
	err := transformShipmentEnvironmentToTerraform(shipmentEnv, d)
	if err != nil {
		return err
	}

	return nil
}

//make updates to remote resource (use shipit bulk and trigger)
func resourceHarborShipmentEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*harborMeta).auth
	shipmentName, env := idParts(d.Id())

	//lookup the shipment in order to get the group/envvars (required for bulk creating env)
	shipment := GetShipment(auth.Username, auth.Token, shipmentName)
	if shipment == nil {
		return errors.New("shipment not found")
	}

	//is there an existing image to apply?
	existingImage := ""
	if meta.(*harborMeta).state["image"] != nil {
		existingImage = meta.(*harborMeta).state["image"].(string)
		if Verbose {
			log.Println("existing image = " + existingImage)
		}
	}

	//transform tf resource data into shipit model
	shipmentEnv, err := transformTerraformToShipmentEnvironment(d, existingImage, shipment.Group, shipment.EnvVars)
	if err != nil {
		return err
	}

	//add auth
	shipmentEnv.Username = auth.Username
	shipmentEnv.Token = auth.Token

	//debug print json
	if Verbose {
		b, _ := json.MarshalIndent(shipmentEnv, "\t", "\t")
		log.Println(string(b))
	}

	//post new shipment/environment
	SaveNewShipmentEnvironment(auth.Username, auth.Token, *shipmentEnv)

	//trigger shipment
	success, messages := Trigger(shipmentName, env)
	if !success {
		failureMessage := ""
		for _, m := range messages {
			failureMessage += m + "\n"
		}
		return fmt.Errorf("trigger failed: %v", failureMessage)
	}

	return nil
}

//populate a terraform ResourceData from a shipit ShipmentEnvironment
func transformShipmentEnvironmentToTerraform(shipmentEnv *ShipmentEnvironment, d *schema.ResourceData) error {

	//set attributes
	d.Set("environment", shipmentEnv.Name)
	d.Set("monitored", shipmentEnv.EnableMonitoring)

	provider := ec2Provider(shipmentEnv.Providers)
	d.Set("barge", provider.Barge)
	d.Set("replicas", provider.Replicas)

	//[]map[string]interface{}
	containers := make([]map[string]interface{}, len(shipmentEnv.Containers))
	for i, container := range shipmentEnv.Containers {
		c := make(map[string]interface{})
		c["name"] = container.Name
		containers[i] = c

		//ports
		ports := make([]map[string]interface{}, len(shipmentEnv.Containers[i].Ports))
		for j, port := range shipmentEnv.Containers[i].Ports {
			p := make(map[string]interface{})
			p["primary"] = port.Primary
			p["value"] = port.Value
			p["public_port"] = port.PublicPort
			p["public"] = port.PublicVip
			p["external"] = port.External
			p["protocol"] = port.Protocol
			p["healthcheck"] = port.Healthcheck
			p["enable_proxy_protocol"] = port.EnableProxyProtocol
			p["ssl_arn"] = port.SslArn
			p["ssl_management_type"] = port.SslManagementType
			//TODO:
			//p["private_key"] = port.
			//p["public_key_certificate"] = port.
			//p["certificate_chain"] = port.
			ports[j] = p
		}
		c["port"] = ports
	}
	err := d.Set("container", containers)
	if err != nil {
		return err
	}

	//read hc settings from primary first container's port

	return nil
}

//populate a shipit ShipmentEnvironment from a terraform ResourceData
func transformTerraformToShipmentEnvironment(d *schema.ResourceData, existingImage string, group string, shipmentEnvVars []EnvVarPayload) (*ShipmentEnvironment, error) {

	result := ShipmentEnvironment{
		ParentShipment: ParentShipment{
			Name:    d.Get("shipment").(string),
			Group:   group,
			EnvVars: shipmentEnvVars,
		},
		Name:             d.Get("environment").(string),
		EnableMonitoring: d.Get("monitored").(bool),
	}

	//add default ec2 provider
	provider := ProviderPayload{
		Name:     providerEc2,
		Barge:    d.Get("barge").(string),
		Replicas: d.Get("replicas").(int),
		EnvVars:  make([]EnvVarPayload, 0),
	}
	result.Providers = append(result.Providers, provider)

	//map containers
	containersResource := d.Get("container") //[]map[string]interface{}
	if containers, ok := containersResource.([]interface{}); ok && len(containers) > 0 {
		result.Containers = make([]ContainerPayload, len(containers))
		for i, c := range containers {
			ctr := c.(map[string]interface{})

			//container properties
			result.Containers[i].Name = ctr["name"].(string)
			result.Containers[i].Image = existingImage

			//map ports
			hcTimeout := d.Get("healthcheck_timeout").(int)
			hcInterval := d.Get("healthcheck_interval").(int)

			if portsResource, ok := ctr["port"].([]interface{}); ok && len(portsResource) > 0 {
				currentContainer := &result.Containers[i]
				currentContainer.Ports = make([]PortPayload, len(portsResource))
				for j, port := range portsResource {
					if portMap, ok := port.(map[string]interface{}); ok {
						p := &currentContainer.Ports[j]

						//port configuration
						p.Name = "PORT"
						if j > 0 {
							p.Name = fmt.Sprintf("%v_%v", p.Name, j)
						}
						p.Protocol = portMap["protocol"].(string)
						p.Value = portMap["value"].(int)
						p.Healthcheck = portMap["healthcheck"].(string)
						p.PublicPort = portMap["public_port"].(int)
						p.External = portMap["external"].(bool)
						p.EnableProxyProtocol = portMap["enable_proxy_protocol"].(bool)
						p.External = portMap["external"].(bool)
						p.Primary = portMap["primary"].(bool)
						p.PublicVip = portMap["public"].(bool)
						p.SslArn = portMap["ssl_arn"].(string)
						p.SslManagementType = portMap["ssl_management_type"].(string)

						//use default backend if existing image is not supplied
						if existingImage == "" {
							if Verbose {
								log.Println("using default backend")
							}

							result.Containers[i].Image = fmt.Sprintf("%v:%v", defaultBackendImageName, defaultBackendImageVersion)

							//add PORT env var required by default backend image
							result.Containers[i].EnvVars = make([]EnvVarPayload, 1)
							result.Containers[i].EnvVars[0].Name = "PORT"
							result.Containers[i].EnvVars[0].Value = strconv.Itoa(portMap["value"].(int))
						}

						//map hc settings down to all ports
						p.HealthcheckTimeout = &hcTimeout
						p.HealthcheckInterval = &hcInterval

					} else {
						return nil, errors.New("port is not a map[string]interface{}")
					}
				}
			} else {
				return nil, errors.New("at least 1 port is required")
			}
		} //iterate containers

	} else {
		return nil, errors.New("at least 1 container is required")
	}

	return &result, nil
}
