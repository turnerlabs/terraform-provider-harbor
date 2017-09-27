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

const defaultBackendImage = "quay.io/turner/turner-defaultbackend:0.1.0"

func resourceHarborShipmentEnv() *schema.Resource {
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

func transformTerraformToShipmentEnvironment(d *schema.ResourceData, group string, shipmentEnvVars []EnvVarPayload) (*ShipmentEnvironment, error) {

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
			result.Containers[i].Image = defaultBackendImage
			result.Containers[i].EnvVars = make([]EnvVarPayload, 1)

			//add PORT env var to configure default backend
			result.Containers[i].EnvVars[0].Name = "PORT"

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
						result.Containers[i].EnvVars[0].Value = strconv.Itoa(p.Value)
						p.PublicPort = portMap["public_port"].(int)
						p.External = portMap["external"].(bool)
						p.EnableProxyProtocol = portMap["enable_proxy_protocol"].(bool)
						p.External = portMap["external"].(bool)
						//p.Healthcheck = portMap["healthcheck"].(string)
						p.Healthcheck = "/healthz"
						p.Primary = portMap["primary"].(bool)
						p.PublicVip = portMap["public"].(bool)
						p.SslArn = portMap["ssl_arn"].(string)
						p.SslManagementType = portMap["ssl_management_type"].(string)

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

func resourceHarborShipmentEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*Auth)
	shipmentName := d.Get("shipment").(string)
	environment := d.Get("environment").(string)

	//lookup the shipment in order to get the group/envvars (required for bulk creating env)
	shipment := GetShipment(auth.Username, auth.Token, shipmentName)
	if shipment == nil {
		return errors.New("shipment not found")
	}

	//transform tf resource data into shipit model
	shipmentEnv, err := transformTerraformToShipmentEnvironment(d, shipment.Group, shipment.EnvVars)
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

func getShipmentEnv(id string) (string, string) {
	parts := strings.Split(id, "::")
	return parts[0], parts[1]
}

func resourceHarborShipmentEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	auth := meta.(*Auth)
	shipment, env := getShipmentEnv(d.Id())

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

	return nil
}

func resourceHarborShipmentEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	// uri := fullyQualifiedInfrastructureURI("shipment/")
	// res, _, err := gorequest.New().Put(uri).
	// 	Set("x-username", meta.(*Auth).Username).
	// 	Set("x-token", meta.(*Auth).Token).
	// 	End()
	// if err != nil {
	// 	return err[0]
	// }
	// if res.StatusCode != 200 {
	// 	return errors.New(uri + " create api returned " + strconv.Itoa(res.StatusCode))
	// }
	return nil
}

func resourceHarborShipmentEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	// if d.Id() == "" {
	// 	return nil
	// }

	// //todo: lookup shipment/env by id (shipment::env)

	// auth := meta.(*Auth)
	// GetShipmentEnvironment(auth.Username, auth.Token, d.Id())

	// uri := fullyQualifiedInfrastructureURI(d.Id())
	// res, body, err := gorequest.New().Get(uri).EndBytes()
	// if err != nil {
	// 	return err[0]
	// }
	// if res.StatusCode == 404 {
	// 	return nil
	// } else if res.StatusCode != 200 {
	// 	return errors.New("get environment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	// }

	// var result ShipmentEnv
	// unmarshalErr := json.Unmarshal(body, &result)
	// if unmarshalErr != nil {
	// 	return unmarshalErr
	// }

	// d.Set("environment", result.Environment)
	// d.Set("barge", result.Barge)
	// d.Set("replicas", result.Replicas)

	return nil
}
