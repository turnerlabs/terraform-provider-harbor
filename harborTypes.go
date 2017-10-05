package main

import "time"

// AuthRequest represents an authentication request
type AuthRequest struct {
	User string `json:"username,omitempty"`
	Pass string `json:"password,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success bool
	Token   string
}

// HarborCompose represents a harbor-compose.yml file
type HarborCompose struct {
	Shipments map[string]ComposeShipment `yaml:"shipments"`
}

// ComposeShipment represents a harbor shipment in a harbor-compose.yml file
type ComposeShipment struct {
	Env                        string            `yaml:"env"`
	Barge                      string            `yaml:"barge"`
	Containers                 []string          `yaml:"containers"`
	Replicas                   int               `yaml:"replicas"`
	Group                      string            `yaml:"group"`
	Property                   string            `yaml:"property"`
	Project                    string            `yaml:"project"`
	Product                    string            `yaml:"product"`
	Environment                map[string]string `yaml:"environment,omitempty"`
	IgnoreImageVersion         bool              `yaml:"ignoreImageVersion,omitempty"`
	EnableMonitoring           *bool             `yaml:"enableMonitoring,omitempty"`
	HealthcheckTimeoutSeconds  *int              `yaml:"healthcheckTimeoutSeconds,omitempty"`
	HealthcheckIntervalSeconds *int              `yaml:"healthcheckIntervalSeconds,omitempty"`
}

// DockerCompose represents a docker-compose.yml file (only used for writing via generate/init)
type DockerCompose struct {
	Version  string                           `yaml:"version"`
	Services map[string]*DockerComposeService `yaml:"services"`
}

// DockerComposeService represents a container (only used for writing via generate/init)
type DockerComposeService struct {
	Build       string            `yaml:"build,omitempty"`
	Image       string            `yaml:"image,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	EnvFile     []string          `yaml:"env_file,omitempty"`
}

// ShipmentEnvironment represents a shipment/environment combination
type ShipmentEnvironment struct {
	Username         string             `json:"username"`
	Token            string             `json:"token"`
	Name             string             `json:"name"`
	EnvVars          []EnvVarPayload    `json:"envVars,omitempty"`
	Ports            []PortPayload      `json:"ports,omitempty"`
	Containers       []ContainerPayload `json:"containers,omitempty"`
	Providers        []ProviderPayload  `json:"providers,omitempty"`
	ParentShipment   ParentShipment     `json:"parentShipment"`
	BuildToken       string             `json:"buildToken,omitempty"`
	EnableMonitoring bool               `json:"enableMonitoring"`
}

// The ParentShipment of the shipmentModel
type ParentShipment struct {
	Name    string          `json:"name"`
	EnvVars []EnvVarPayload `json:"envVars"`
	Group   string          `json:"group"`
}

// EnvVarPayload represents EnvVar
type EnvVarPayload struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
}

// PortPayload represents a port
type PortPayload struct {
	Name                string `json:"name"`
	Value               int    `json:"value,omitempty"`
	Protocol            string `json:"protocol,omitempty"`
	Healthcheck         string `json:"healthcheck,omitempty"`
	Primary             bool   `json:"primary"`
	External            bool   `json:"external,omitempty"`
	PublicVip           bool   `json:"public_vip,omitempty"`
	PublicPort          int    `json:"public_port,omitempty"`
	EnableProxyProtocol bool   `json:"enable_proxy_protocol,omitempty"`
	SslArn              string `json:"ssl_arn,omitempty"`
	SslManagementType   string `json:"ssl_management_type,omitempty"`
	HealthcheckTimeout  *int   `json:"healthcheck_timeout,omitempty"`
	HealthcheckInterval *int   `json:"healthcheck_interval,omitempty"`
}

// ContainerPayload represents a container payload
type ContainerPayload struct {
	Name    string          `json:"name,omitempty"`
	Image   string          `json:"image,omitempty"`
	EnvVars []EnvVarPayload `json:"envVars,omitempty"`
	Ports   []PortPayload   `json:"ports,omitempty"`
}

// ProviderPayload represents a provider payload
type ProviderPayload struct {
	Name     string          `json:"name"`
	Replicas int             `json:"replicas"`
	EnvVars  []EnvVarPayload `json:"envVars,omitempty"`
	Barge    string          `json:"barge,omitempty"`
}

// TriggerResponseSingle is the payload returned from the trigger api
type TriggerResponseSingle struct {
	Message string `json:"message,omitempty"`
}

// TriggerResponseMultiple is the payload returned from the trigger api
type TriggerResponseMultiple struct {
	Messages []string `json:"message,omitempty"`
}

// ContainerStatusOutput represents an object that can be written to stdout and formatted
type ContainerStatusOutput struct {
	ID        string
	Image     string
	Started   string
	Status    string
	Restarts  string
	LastState string
}

//ShipmentStatusOutput represents an object that can be written to stdout and formatted
type ShipmentStatusOutput struct {
	Shipment    string
	Environment string
	Barge       string
	Status      string
	Containers  string
	Replicas    string
	Endpoint    string
}

// CatalogitContainer is what gets sent to catalog to post a new image
type CatalogitContainer struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
}

// DeployRequest represents a request to deploy a shipment/container to an environment
type DeployRequest struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
	Catalog bool   `json:"catalog"`
}

// UpdateShipmentEnvironmentRequest represents a request to update a shipment/environment
type UpdateShipmentEnvironmentRequest struct {
	EnableMonitoring bool `json:"enableMonitoring"`
}

// UpdatePortRequest represents a request to update a port
type UpdatePortRequest struct {
	Name                string `json:"name"`
	HealthcheckTimeout  *int   `json:"healthcheck_timeout,omitempty"`
	HealthcheckInterval *int   `json:"healthcheck_interval,omitempty"`
}

// HelmitContainer represents a single container instance in harbor
type HelmitContainer struct {
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	Image     string   `json:"image"`
	Logstream string   `json:"log_stream"`
	Logs      []string `json:"logs"`
}

// HelmitReplica represents a single running replica in harbor
type HelmitReplica struct {
	Host       string            `json:"host"`
	Provider   string            `json:"provider"`
	Containers []HelmitContainer `json:"containers"`
}

// HelmitResponse represents a response from helmit
type HelmitResponse struct {
	Error    bool            `json:"error"`
	Replicas []HelmitReplica `json:"replicas"`
}

//ShipmentStatus represents the deployed status of a shipment
type ShipmentStatus struct {
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	Status    struct {
		Phase      string `json:"phase"`
		Containers []struct {
			ID        string                        `json:"id"`
			Image     string                        `json:"image"`
			Ready     bool                          `json:"ready"`
			Restarts  int                           `json:"restarts"`
			State     map[string]ContainerState     `json:"state"`
			Status    string                        `json:"status"`
			LastState map[string]ContainerLastState `json:"lastState"`
		} `json:"containers"`
	} `json:"status"`
	AverageRestarts float32 `json:"averageRestarts"`
}

// ContainerState represents a particular state of a container
type ContainerState struct {
	StartedAt time.Time `json:"startedAt"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
}

// ContainerLastState represents the last state of a container
type ContainerLastState struct {
	ExitCode    int       `json:"exitCode"`
	Reason      string    `json:"reason"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	ContainerID string    `json:"containerID"`
}

// Shipment reprents a top-level shipment
type Shipment struct {
	Name    string          `json:"name,omitempty"`
	Group   string          `json:"group,omitempty"`
	EnvVars []EnvVarPayload `json:"envVars,omitempty"`
}

type getLoadBalancerStatusResponse struct {
	LoadBalancers []struct {
		LoadBalancerArn       string `json:"load_balancer_arn"`
		DNSName               string `json:"dns_name"`
		CanonicalHostedZoneID string `json:"canonical_hosted_zone_id"`
		CreatedTime           string `json:"created_time"`
		LoadBalancerName      string `json:"load_balancer_name"`
		Scheme                string `json:"scheme"`
		VpcID                 string `json:"vpc_id"`
		State                 struct {
			Code string `json:"code"`
		} `json:"state"`
		Type              string `json:"type"`
		AvailabilityZones []struct {
			ZoneName string `json:"zone_name"`
			SubnetID string `json:"subnet_id"`
		} `json:"availability_zones"`
		SecurityGroups []string `json:"security_groups"`
		IPAddressType  string   `json:"ip_address_type"`
	} `json:"load_balancers"`
	LbName string `json:"lb_name"`
}
