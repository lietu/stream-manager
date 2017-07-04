package lametric

import (
	"encoding/json"
	"github.com/lietu/stream-manager/config"
	"log"
	"fmt"
	"net/http"
	"bytes"
	"io/ioutil"
)

type LaMetric struct {
	config *config.LaMetric
}

func (l *LaMetric) request(notification_type string, text string) {
	var c *config.LaMetricApp
	if notification_type == "follower" {
		c = l.config.Follower
	} else if notification_type == "host" {
		c = l.config.Host
	} else if notification_type == "bits" {
		c = l.config.Bits
	} else if notification_type == "subscriber" {
		c = l.config.Subscriber
	} else {
		return
	}

	if c == nil {
		return
	}

	f := requestFrame{}
	f.Text = text
	f.Icon = c.Icon
	f.Index = 0

	r := request{}
	r.Frames = []requestFrame{f}

	data, err := json.Marshal(r)
	if err != nil {
		log.Printf("Error formatting LaMetric request: %s", err)
		return
	}

	req, err := http.NewRequest("POST", c.Url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating LaMetric request: %s", err)
		return
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Access-Token", c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making LaMetric request: %s", err)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("LaMetric request %s: %s", resp.Status, body)
}

func (l *LaMetric) Host(service string, name string) {
	l.request("host", fmt.Sprintf("%s host from %s", service, name))
}

func (l *LaMetric) Bits(name string, bits int, message string) {
	l.request("bits", fmt.Sprintf("%s cheered for %d bits", name, bits))
}

func (l *LaMetric) Follower(service string, name string) {
	l.request("follower", fmt.Sprintf("New %s follower: %s", service, name))
}

func (l *LaMetric) Subscriber(service string, name string, tier string, months string) {
	l.request("subscriber", fmt.Sprintf("%s did a %s sub on %s for %s months in a row", name, tier, service, months))
}

func New(config *config.LaMetric) *LaMetric {
	l := LaMetric{}
	l.config = config
	return &l
}
