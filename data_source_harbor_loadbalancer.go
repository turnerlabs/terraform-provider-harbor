package main

import (
	"math/rand"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceHarborLoadbalancer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHarborLoadbalancerRead,

		Schema: map[string]*schema.Schema{
			"shipment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			//attributes
			"dns_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRandomID() string {
	b := make([]rune, 30)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func dataSourceHarborLoadbalancerRead(d *schema.ResourceData, meta interface{}) error {
	harborMeta := meta.(*harborMeta)
	auth := harborMeta.auth

	writeMetric(metricHarborLoadbalancerRead, auth.Username)
	d.SetId(generateRandomID())

	shipment := d.Get("shipment").(string)
	environment := d.Get("environment").(string)

	//query harbor for the lb status
	result, err := getLoadBalancerStatus(shipment, environment)
	if err != nil {
		writeMetricError(metricHarborLoadbalancerRead, auth.Username, err)
		return err
	}

	//set computed attributes
	d.Set("dns_name", result.DNSName)
	d.Set("name", result.Name)
	d.Set("type", result.Type)
	d.Set("arn", result.ARN)
	d.Set("dns_name", result.DNSName)
	d.Set("hosted_zone_id", result.CanonicalHostedZoneID)

	return nil
}
