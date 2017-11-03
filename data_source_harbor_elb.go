package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/parnurzeal/gorequest"
)

func dataSourceHarborElb() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHarborElbRead,

		Schema: map[string]*schema.Schema{
			"shipment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"dns_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type loadBalancerStatus struct {
	Name string `json:"lb_name"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRandomID() string {
	b := make([]rune, 30)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func dataSourceHarborElbRead(d *schema.ResourceData, meta interface{}) error {
	writeMetric(metricHarborElbRead)
	d.SetId(generateRandomID())

	shipment := d.Get("shipment").(string)
	environment := d.Get("environment").(string)

	//query harbor for the lb_name
	auth := meta.(*Auth)
	uri := fmt.Sprintf("http://harbor-trigger.services.dmtio.net/loadbalancer/status/%v/%v/ec2", shipment, environment)
	res, body, err := gorequest.New().Get(uri).
		Set("x-username", auth.Username).
		Set("x-token", auth.Token).
		EndBytes()

	if err != nil {
		writeMetricError(metricHarborElbRead, err[0])
		return err[0]
	}
	if res.StatusCode != 200 {
		newErr := errors.New("get load balancer status api returned " + strconv.Itoa(res.StatusCode) + " for " + uri)
		writeMetricError(metricHarborElbRead, newErr)
		return newErr
	}

	var lb loadBalancerStatus
	unmarshalErr := json.Unmarshal(body, &lb)
	if unmarshalErr != nil {
		writeMetricError(metricHarborElbRead, unmarshalErr)
		return unmarshalErr
	}

	//query aws for the lb dns name
	session := session.Must(session.NewSession())
	elbconn := elb.New(session)

	// Retrieve the ELB properties for updating the state
	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(lb.Name)},
	}

	describeResp, awserr := elbconn.DescribeLoadBalancers(describeElbOpts)
	if err != nil {
		newErr := fmt.Errorf("Error retrieving ELB: %s", awserr)
		writeMetricError(metricHarborElbRead, newErr)
		return newErr
	}
	if describeResp == nil || describeResp.LoadBalancerDescriptions == nil || len(describeResp.LoadBalancerDescriptions) == 0 {
		newErr := errors.New("load balancer not found")
		writeMetricError(metricHarborElbRead, newErr)
		return newErr
	}

	//output task definition json
	d.Set("dns_name", describeResp.LoadBalancerDescriptions[0].DNSName)

	return nil
}
