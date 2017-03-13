package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/parnurzeal/gorequest"
)

func create(uriResource string, auth *Auth, data interface{}) error {

	uri := fullyQualifiedURI(uriResource)
	res, _, err := gorequest.New().Post(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("create api returned " + strconv.Itoa(res.StatusCode))
	}

	return nil
}

func update(uriResource string, auth *Auth, data interface{}) error {

	uri := fullyQualifiedURI(uriResource)
	res, _, err := gorequest.New().Put(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		Send(data).
		End()

	if err != nil {
		return err[0]
	}

	if res.StatusCode != 200 {
		return errors.New("update api returned " + strconv.Itoa(res.StatusCode) + " for " + uriResource)
	}

	return nil
}

func delete(uriResource string, auth *Auth) error {

	uri := fullyQualifiedURI(uriResource)
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
		return errors.New("delete api returned " + strconv.Itoa(res.StatusCode) + " for " + uriResource)
	}

	return nil
}

func parseContainerResourceURI(uri string) (string, string, string) {
	parts := strings.Split(uri, "/")
	shipmentEnvURI := strings.Join(parts[:4], "/")
	containerName := parts[5]
	resourceName := parts[len(parts)-1]

	return shipmentEnvURI, containerName, resourceName
}

func readContainer(shipmentEnvironmentURI string, containerName string, auth *Auth) (*containerPayload, error) {

	//fetch the shipment environment
	uri := fullyQualifiedURI(shipmentEnvironmentURI)
	res, body, err := gorequest.New().Get(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		EndBytes()
	if err != nil {
		return nil, err[0]
	}
	if res.StatusCode != 200 {
		return nil, errors.New("get environment api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
	}

	var result shipmentEnvironment
	unmarshalErr := json.Unmarshal(body, &result)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	//try to find container in environment resource by container name
	if len(result.Containers) == 0 {
		return nil, nil
	}

	var matchingContainer *containerPayload
	for _, container := range result.Containers {
		if container.Name == containerName {
			matchingContainer = &container
			break
		}
	}

	if matchingContainer == nil {
		return nil, nil
	}

	return matchingContainer, nil
}
