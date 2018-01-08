package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
)

type metric struct {
	Source string `json:"source,omitempty"`
	Action string `json:"action,omitempty"`
	Error  string `json:"error,omitempty"`
	OS     string `json:"os,omitempty"`
	Arch   string `json:"arch,omitempty"`
}

const (
	telemetryURI = "https://telemetry.harbor.turnerlabs.io/v1/api/metric"

	metricShipmentCreate = "shipment.create"
	metricShipmentUpdate = "shipment.update"
	metricShipmentDelete = "shipment.delete"
	metricShipmentImport = "shipment.import"

	metricEnvCreate = "env.create"
	metricEnvUpdate = "env.update"
	metricEnvDelete = "env.delete"
	metricEnvImport = "env.import"

	metricHarborLoadbalancerRead = "harbor_loadbalancer.read"
)

func writeMetric(action string) {
	writeMetricErrorString(action, "")
}

func writeMetricError(action string, err error) {
	writeMetricErrorString(action, err.Error())
}

func writeMetricErrorString(action string, err string) {

	// HARBOR_TELEMETRY=0 disables telemetry
	if setting := os.Getenv("HARBOR_TELEMETRY"); setting != "0" {

		m := metric{
			Source: "terraform-provider-harbor",
			Action: action,
			Error:  err,
			OS:     runtime.GOOS,
			Arch:   runtime.GOARCH,
		}

		if Verbose {
			log.Println("posting telemetry data")
		}

		//talk to the server in the background to keep things moving
		go postTelemetryData(m)
	}
}

func postTelemetryData(m metric) {
	json, _ := json.Marshal(m)
	req, err := http.NewRequest("POST", telemetryURI, bytes.NewBuffer(json))
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
