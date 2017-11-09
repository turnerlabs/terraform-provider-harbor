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

const (
	defaultBackendImageName    = "quay.io/turner/turner-defaultbackend"
	defaultBackendImageVersion = "0.2.0"
)

func resourceHarborShipmentEnv() *schema.Resource {
	return &schema.Resource{
		Create: resourceHarborShipmentEnvironmentCreate,
		Read:   resourceHarborShipmentEnvironmentRead,
		Update: resourceHarborShipmentEnvironmentUpdate,
		Delete: resourceHarborShipmentEnvironmentDelete,
		Exists: resourceHarborShipmentEnvironmentExists,
		Importer: &schema.ResourceImporter{
			State: resourceHarborShipmentEnvironmentImport,
		},

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
			"log_shipping": &schema.Schema{
				Description: "Configure harbor log shipping",
				Optional:    true,
				MinItems:    0,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Description: "Which provider to use to send logs. Possible values are: logzio, elasticsearch, aws-elasticsearch, sqs, loggly",
							Type:        schema.TypeString,
							Required:    true,
						},
						"endpoint": {
							Description: "provider's endpoint",
							Type:        schema.TypeString,
							Required:    true,
						},
						"aws_access_key": {
							Description: "aws access key (required by aws-elasticsearch and sqs providers)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"aws_secret_key": {
							Description: "aws secret key (required by aws-elasticsearch and sqs providers)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"aws_region": {
							Description: "aws region (required by aws-elasticsearch provider)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"aws_elasticsearch_domain_name": {
							Description: "elastic search domain name (required by aws-elasticsearch provider)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"sqs_queue_name": {
							Description: "sqs queue name (required by sqs provider)",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"annotations": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"container": &schema.Schema{
				Description: "The list of containers for this shipment environment",
				Optional:    true,
				ForceNew:    true,
				MinItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"port": {
							Optional: true,
							ForceNew: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"healthcheck": &schema.Schema{
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
										ForceNew: true,
									},
									"healthcheck_timeout": &schema.Schema{
										Type:     schema.TypeInt,
										Optional: true,
										Default:  1,
									},
									"healthcheck_interval": &schema.Schema{
										Type:     schema.TypeInt,
										Optional: true,
										Default:  10,
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
			"dns_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
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
	harborMeta := meta.(*harborMeta)
	auth := harborMeta.auth

	shipmentName := d.Get("shipment").(string)
	environment := d.Get("environment").(string)

	//lookup the shipment in order to get the group/envvars (required for bulk creating env)
	shipment := GetShipment(auth.Username, auth.Token, shipmentName)
	if shipment == nil {
		return errors.New("shipment not found")
	}

	//transform tf resource data into shipit model
	shipmentEnv, err := transformTerraformToShipmentEnvironment(d, harborMeta.existingShipmentEnvironment, shipment.Group, shipment.EnvVars)
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

	//validate before saving
	err = validateShipmentEnvironment(shipmentEnv)
	if err != nil {
		return err
	}

	//save shipment/environment
	writeMetric(metricEnvCreate)
	SaveShipmentEnvironment(auth.Username, auth.Token, *shipmentEnv)

	//trigger shipment
	success, messages := Trigger(shipmentName, environment)
	if !success {
		failureMessage := ""
		for _, m := range messages {
			failureMessage += m + "\n"
		}
		newErr := fmt.Errorf("trigger failed: %v", failureMessage)
		writeMetricError(metricEnvCreate, newErr)
		return newErr
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
		newErr := errors.New("no load balancers")
		writeMetricError(metricEnvCreate, newErr)
		return newErr
	}

	//output id
	d.SetId(fmt.Sprintf("%s::%s", shipmentEnv.ParentShipment.Name, shipmentEnv.Name))

	//output attributes
	d.Set("dns_name", fmt.Sprintf("%v.%v.services.ec2.dmtio.net", shipmentName, environment))
	d.Set("lb_name", lbStatus.LoadBalancers[0].LoadBalancerName)
	d.Set("lb_type", lbStatus.LoadBalancers[0].Type)
	d.Set("lb_arn", lbStatus.LoadBalancers[0].LoadBalancerArn)
	d.Set("lb_dns_name", lbStatus.LoadBalancers[0].DNSName)
	d.Set("lb_hosted_zone_id", lbStatus.LoadBalancers[0].CanonicalHostedZoneID)

	return nil
}

func validateShipmentEnvironment(shipmentEnv *ShipmentEnvironment) error {

	// - only 1 port is primary
	// - only 1 healthcheck per container
	// - port.public_port must be unique per env

	primaryPorts := 0
	publicPorts := make(map[int]int)
	for _, container := range shipmentEnv.Containers {
		hcPorts := 0
		for _, port := range container.Ports {
			if port.Healthcheck != "" {
				hcPorts++
				if hcPorts > 1 {
					return fmt.Errorf("Container '%v' must have only 1 healthcheck port. Please remove the healthcheck from the other ports.", container.Name)
				}
			}
			if port.Primary {
				primaryPorts++
				if primaryPorts > 1 {
					return errors.New("must have exactly 1 primary container. Add 'primary = false' to non-primary containers")
				}
			}
			if publicPorts[port.PublicPort] > 0 {
				return fmt.Errorf("public_port '%v' must be unique", port.PublicPort)
			}
			publicPorts[port.PublicPort] = port.PublicPort
		}
	}

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

	writeMetric(metricEnvDelete)

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

	//store the existing ShipmentEnvironment in meta state
	//so that create can re-attach user images
	harborMeta.existingShipmentEnvironment = shipmentEnv

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

func resourceHarborShipmentEnvironmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	writeMetric(metricEnvImport)

	//lookup and set the arguments
	auth := meta.(*harborMeta).auth
	shipment, env := idParts(d.Id())
	shipmentEnv := GetShipmentEnvironment(auth.Username, auth.Token, shipment, env)
	if shipmentEnv == nil {
		newErr := errors.New("shipment/environment doesn't exist")
		writeMetricError(metricEnvImport, newErr)
		return nil, newErr
	}

	//transform shipit model back to terraform
	err := transformShipmentEnvironmentToTerraform(shipmentEnv, d)
	if err != nil {
		writeMetricError(metricEnvImport, err)
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

//make updates to remote resource (use shipit bulk and trigger)
func resourceHarborShipmentEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	harborMeta := meta.(*harborMeta)
	auth := harborMeta.auth
	shipmentName, env := idParts(d.Id())

	//lookup existing shipment/env
	shipmentEnv := GetShipmentEnvironment(auth.Username, auth.Token, shipmentName, env)
	if shipmentEnv == nil {
		return errors.New("shipment/environment doesn't exist")
	}

	//transform tf resource data into shipit model
	shipmentEnv, err := transformTerraformToShipmentEnvironment(d, shipmentEnv, shipmentEnv.ParentShipment.Group, shipmentEnv.ParentShipment.EnvVars)
	if err != nil {
		return err
	}

	shipmentEnv.Username = auth.Username
	shipmentEnv.Token = auth.Token

	//debug print json
	if Verbose {
		b, _ := json.MarshalIndent(shipmentEnv, "\t", "\t")
		log.Println(string(b))
	}

	//validate before saving
	err = validateShipmentEnvironment(shipmentEnv)
	if err != nil {
		return err
	}

	//save shipment/environment
	writeMetric(metricEnvUpdate)
	SaveShipmentEnvironment(auth.Username, auth.Token, *shipmentEnv)

	//trigger shipment
	success, messages := Trigger(shipmentName, env)
	if !success {
		failureMessage := ""
		for _, m := range messages {
			failureMessage += m + "\n"
		}
		newErr := fmt.Errorf("trigger failed: %v", failureMessage)
		writeMetricError(metricEnvUpdate, newErr)
		return newErr
	}

	return nil
}

//populate a terraform ResourceData from a shipit ShipmentEnvironment
func transformShipmentEnvironmentToTerraform(shipmentEnv *ShipmentEnvironment, d *schema.ResourceData) error {

	//set attributes
	d.Set("shipment", shipmentEnv.ParentShipment.Name)
	d.Set("environment", shipmentEnv.Name)
	d.Set("monitored", shipmentEnv.EnableMonitoring)

	provider := ec2Provider(shipmentEnv.Providers)
	d.Set("barge", provider.Barge)
	d.Set("replicas", provider.Replicas)

	// annotations
	annotations := make(map[string]string, len(shipmentEnv.Annotations))
	for _, anno := range shipmentEnv.Annotations {
		annotations[anno.Key] = anno.Value
	}
	annoErr := d.Set("annotations", annotations)
	if annoErr != nil {
		return annoErr
	}

	//log shipping
	envvar := findEnvVar(envVarNameShipLogs, shipmentEnv.EnvVars)
	if envvar != (EnvVarPayload{}) {
		log.Println("translating log shipping env vars")

		logShipping := make([]map[string]interface{}, 1)
		logShippingConfig := make(map[string]interface{})
		logShippingConfig["provider"] = envvar.Value

		if envvar = findEnvVar(envVarNameLogsEndpoint, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["endpoint"] = envvar.Value
		}

		if envvar = findEnvVar(envVarNameDomainName, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["aws_elasticsearch_domain_name"] = envvar.Value
		}

		if envvar := findEnvVar(envVarNameRegion, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["aws_region"] = envvar.Value
		}

		if envvar = findEnvVar(envVarNameAccessKey, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["aws_access_key"] = envvar.Value
		}

		if envvar = findEnvVar(envVarNameSecretKey, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["aws_secret_key"] = envvar.Value
		}

		if envvar = findEnvVar(envVarNameQueueName, shipmentEnv.EnvVars); envvar != (EnvVarPayload{}) {
			logShippingConfig["sqs_queue_name"] = envvar.Value
		}

		logShipping[0] = logShippingConfig
		err := d.Set("log_shipping", logShipping)
		if err != nil {
			return err
		}
	} else { //SHIP_LOGS not found
		//remove tf config
		err := d.Set("log_shipping", nil)
		if err != nil {
			return err
		}
	}

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
			p["value"] = port.Value
			p["public_port"] = port.PublicPort
			p["public"] = port.PublicVip
			p["external"] = port.External
			p["protocol"] = port.Protocol
			p["enable_proxy_protocol"] = port.EnableProxyProtocol
			p["ssl_arn"] = port.SslArn
			p["ssl_management_type"] = port.SslManagementType
			p["healthcheck"] = port.Healthcheck
			p["healthcheck_timeout"] = *port.HealthcheckTimeout
			p["healthcheck_interval"] = *port.HealthcheckInterval

			//set container as primary since it contains the shipment/env's primary port
			//and there can only be 1 per shipment/env
			if port.Primary {
				c["primary"] = true
			}

			ports[j] = p
		}
		c["port"] = ports
	}
	err := d.Set("container", containers)
	if err != nil {
		return err
	}

	return nil
}

//populate a shipit ShipmentEnvironment from a terraform ResourceData
func transformTerraformToShipmentEnvironment(d *schema.ResourceData, existingShipmentEnvironment *ShipmentEnvironment, group string, shipmentEnvVars []EnvVarPayload) (*ShipmentEnvironment, error) {

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

	//copy over any existing user-defined environment-level env vars
	if existingShipmentEnvironment != nil {
		result.EnvVars = copyUserDefinedEnvVars(existingShipmentEnvironment.EnvVars)
	}

	// annotations
	annotationsResource := d.Get("annotations")
	if annotations, ok := annotationsResource.(map[string]interface{}); ok && len(annotations) > 0 {
		result.Annotations = make([]AnnotationsPayload, len(annotations))

		i := 0
		for k, v := range annotations {
			anno := AnnotationsPayload{
				Key:   k,
				Value: v.(string),
			}

			result.Annotations[i] = anno
			i++
		}
	}

	//translate log_shipping configuration into harbor env vars
	logShippingResource := d.Get("log_shipping") //[]map[string]interface{}
	if logShipping, ok := logShippingResource.([]interface{}); ok && len(logShipping) > 0 {
		log.Println("processing log_shipping")
		ls := logShipping[0].(map[string]interface{})

		//all providers require SHIP_LOGS and LOGS_ENDPOINT
		result.EnvVars = appendEnvVar(result.EnvVars, envVarNameShipLogs, ls["provider"].(string))
		result.EnvVars = appendEnvVar(result.EnvVars, envVarNameLogsEndpoint, ls["endpoint"].(string))

		//provider specific
		switch ls["provider"] {

		case "aws-elasticsearch":
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameDomainName, ls["aws_elasticsearch_domain_name"].(string))
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameRegion, ls["aws_region"].(string))
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameAccessKey, ls["aws_access_key"].(string))
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameSecretKey, ls["aws_secret_key"].(string))

		case "sqs":
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameQueueName, ls["sqs_queue_name"].(string))
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameAccessKey, ls["aws_access_key"].(string))
			result.EnvVars = appendEnvVar(result.EnvVars, envVarNameSecretKey, ls["aws_secret_key"].(string))
		}
	}

	//map containers
	containersResource := d.Get("container") //[]map[string]interface{}
	if containers, ok := containersResource.([]interface{}); ok && len(containers) > 0 {
		result.Containers = make([]ContainerPayload, len(containers))
		for i, c := range containers {
			ctr := c.(map[string]interface{})
			result.Containers[i].Name = ctr["name"].(string)

			//use existing container image, if specified, otherwise use default backend
			useDefaultBackend := true
			if existingShipmentEnvironment != nil {

				//does this container already exist? (user could be adding a new container)
				existingContainer := findContainer(result.Containers[i].Name, existingShipmentEnvironment.Containers)
				if existingContainer.Name != "" {
					if Verbose {
						log.Printf("using existing image/envvars for container: %v\n", result.Containers[i].Name)
					}
					result.Containers[i].Image = existingContainer.Image
					useDefaultBackend = false
				}

				//copy over any existing container env vars
				result.Containers[i].EnvVars = existingContainer.EnvVars
			}

			if useDefaultBackend {
				if Verbose {
					log.Printf("using default backend for container: %v\n", result.Containers[i].Name)
				}

				result.Containers[i].Image = fmt.Sprintf("%v:%v", defaultBackendImageName, defaultBackendImageVersion)

				//add container env vars
				result.Containers[i].EnvVars = make([]EnvVarPayload, 2)

				//configure default backend to use user's port
				result.Containers[i].EnvVars[0].Name = "PORT"
				//value is set later from hc port

				//configure default backend to use user's health check route
				result.Containers[i].EnvVars[1].Name = "HEALTHCHECK"
				//value is set later from hc port
			}

			//map ports
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
						p.PublicPort = portMap["public_port"].(int)
						p.External = portMap["external"].(bool)
						p.EnableProxyProtocol = portMap["enable_proxy_protocol"].(bool)
						p.External = portMap["external"].(bool)
						p.PublicVip = portMap["public"].(bool)
						p.SslArn = portMap["ssl_arn"].(string)
						p.SslManagementType = portMap["ssl_management_type"].(string)

						//healthcheck
						p.Healthcheck = portMap["healthcheck"].(string)
						hcTimeout := portMap["healthcheck_timeout"].(int)
						p.HealthcheckTimeout = &hcTimeout
						hcInterval := portMap["healthcheck_interval"].(int)
						p.HealthcheckInterval = &hcInterval

						//is this the hc port?
						if p.Healthcheck != "" {

							//set container env vars to hc values for default backend
							if useDefaultBackend {
								result.Containers[i].EnvVars[0].Value = strconv.Itoa(p.Value)
								result.Containers[i].EnvVars[1].Value = p.Healthcheck
							}

							//make this port primary if it's an hc port and the container is marked as primary
							if isContainerPrimary := ctr["primary"].(bool); isContainerPrimary {
								p.Primary = true
							}
						}
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
