package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jtacoma/uritemplates"
)

const (
	envVarNameShipLogs     = "SHIP_LOGS"
	envVarNameLogsEndpoint = "LOGS_ENDPOINT"
	envVarNameAccessKey    = "LOGS_ACCESS_KEY"
	envVarNameSecretKey    = "LOGS_SECRET_KEY"
	envVarNameDomainName   = "LOGS_DOMAIN_NAME"
	envVarNameRegion       = "LOGS_REGION"
	envVarNameQueueName    = "LOGS_QUEUE_NAME"
)

func logShippingEnvVars() map[string]string {
	return map[string]string{
		envVarNameShipLogs:     envVarNameShipLogs,
		envVarNameLogsEndpoint: envVarNameLogsEndpoint,
		envVarNameAccessKey:    envVarNameAccessKey,
		envVarNameSecretKey:    envVarNameSecretKey,
		envVarNameDomainName:   envVarNameDomainName,
		envVarNameRegion:       envVarNameRegion,
		envVarNameQueueName:    envVarNameQueueName,
	}
}

func trace(msg string) {
	if Verbose {
		fmt.Println(msg)
	}
}

func check(e error) {
	if e != nil {
		log.Fatal("ERROR: ", e)
	}
}

//find the ec2 provider
func ec2Provider(providers []ProviderPayload) *ProviderPayload {
	for _, provider := range providers {
		if provider.Name == providerEc2 {
			return &provider
		}
	}
	log.Fatal("ec2 provider is missing")
	return nil
}

func appendToFile(file string, lines []string) {
	if _, err := os.Stat(file); err == nil {
		//update
		file, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
		check(err)
		defer file.Close()
		for _, line := range lines {
			_, err = file.WriteString("\n" + line)
			check(err)
		}
	} else {
		//create
		data := ""
		for _, line := range lines {
			data += line + "\n"
		}
		err := ioutil.WriteFile(file, []byte(data), 0644)
		check(err)
	}
}

type tuple struct {
	Item1 string
	Item2 string
}

func param(item1 string, item2 string) tuple {
	return tuple{
		Item1: item1,
		Item2: item2,
	}
}

func buildURI(baseURI string, template string, params ...tuple) string {
	uriTemplate, err := uritemplates.Parse(baseURI + template)
	check(err)
	values := make(map[string]interface{})
	for _, v := range params {
		values[v.Item1] = v.Item2
	}
	uri, err := uriTemplate.Expand(values)
	check(err)
	return uri
}

func findContainer(container string, containers []ContainerPayload) ContainerPayload {
	for _, c := range containers {
		if c.Name == container {
			return c
		}
	}
	return ContainerPayload{}
}

func findEnvVar(name string, envVars []EnvVarPayload) EnvVarPayload {
	for _, e := range envVars {
		if e.Name == name {
			return e
		}
	}
	return EnvVarPayload{}
}

func appendEnvVar(envvars []EnvVarPayload, name string, value string) []EnvVarPayload {
	return append(envvars, EnvVarPayload{
		Name:  name,
		Value: value,
	})
}

//returns a new slice containing user defined (non log-shipping) env vars
func copyUserDefinedEnvVars(envvars []EnvVarPayload) []EnvVarPayload {
	result := []EnvVarPayload{}
	for _, e := range envvars {
		if logShippingEnvVars()[e.Name] == "" {
			result = append(result, e)
		}
	}
	return result
}
