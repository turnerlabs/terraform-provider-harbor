package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type metric struct {
	Source  string `json:"source,omitempty"`
	Action  string `json:"action,omitempty"`
	Error   string `json:"error,omitempty"`
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
	User    string `json:"user,omitempty"`
	Version string `json:"version,omitempty"`
}

const (
	metricShipmentCreate = "shipment.create"
	metricShipmentUpdate = "shipment.update"
	metricShipmentDelete = "shipment.delete"
	metricShipmentImport = "shipment.import"

	metricEnvCreate   = "env.create"
	metricEnvRecreate = "env.recreate"
	metricEnvUpdate   = "env.update"
	metricEnvDelete   = "env.delete"
	metricEnvImport   = "env.import"

	metricHarborLoadbalancerRead = "harbor_loadbalancer.read"
)

func getTelemetryEndpoint() string {
	return GetConfig().TelemetryURI + "/v1/api/metric"
}

func writeMetric(action string, user string) {
	writeMetricErrorString(action, user, "")
}

func writeMetricError(action string, user string, err error) {
	writeMetricErrorString(action, user, err.Error())
}

func getVersion() string {
	//parse version from executable name
	//e.g.: /some/directory/terraform-provider-harbor_v0.5.0
	//take everything after last _
	tmp := strings.Split(os.Args[0], "_")
	version := tmp[len(tmp)-1]
	return version
}

func writeMetricErrorString(action string, user string, err string) {

	// HARBOR_TELEMETRY=0 disables telemetry
	if setting := os.Getenv("HARBOR_TELEMETRY"); setting != "0" {

		m := metric{
			Source:  "terraform-provider-harbor",
			Action:  action,
			Error:   err,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			User:    user,
			Version: getVersion(),
		}

		if Verbose {
			log.Println("posting telemetry data to: " + getTelemetryEndpoint())
		}

		//talk to the server in the background to keep things moving
		go postTelemetryData(m)
	}
}

func postTelemetryData(m metric) {
	json, _ := json.Marshal(m)
	req, err := http.NewRequest("POST", getTelemetryEndpoint(), bytes.NewBuffer(json))
	req.Header.Set("X-key", "0vgKlex4EUckdHYCJq2BPBCyJ5E")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error posting telemetry data: %s\n", err)
		return
	}
	resp.Body.Close()
}
