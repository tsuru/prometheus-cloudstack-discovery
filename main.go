package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/tsuru/prometheus-cloudstack-discovery/cloudstack"
)

func listMachineByProject(c *cloudstack.Client, projectID string, mc chan []string) {
	var machines []string
	defer func() { mc <- machines }()
	params := map[string]string{
		"projectid": projectID,
		"simple":    "true",
	}
	var m cloudstack.ListVirtualMachinesResponse
	err := c.Do("listVirtualMachines", params, &m)
	if err != nil {
		return
	}
	for _, vm := range m.ListVirtualMachinesResponse.VirtualMachine {
		for _, n := range vm.Nic {
			machines = append(machines, n.IpAddress)
		}
	}
}

func listMachines(c *cloudstack.Client) ([]string, error) {
	params := map[string]string{"simple": "true"}
	var projects cloudstack.ListProjectsResponse
	err := c.Do("listProjects", params, &projects)
	if err != nil {
		return nil, err
	}
	mc := make(chan []string)
	for _, p := range projects.ListProjectsResponse.Project {
		go listMachineByProject(c, p.Id, mc)
	}
	var machines []string
	for range projects.ListProjectsResponse.Project {
		machines = append(machines, <-mc...)
	}
	close(mc)
	return machines, err
}

func main() {
	log.SetOutput(ioutil.Discard)
	var (
		address   = flag.String("url", "", "cloudstack api url address")
		apiKey    = flag.String("api-key", "", "cloudstack api key")
		secretKey = flag.String("secret-key", "", "cloudstack secret key")
	)
	flag.Parse()
	c := &cloudstack.Client{
		ApiKey:    url.QueryEscape(*apiKey),
		SecretKey: url.QueryEscape(*secretKey),
		URL:       *address,
	}
	machines, err := listMachines(c)
	if err != nil {
		log.Fatal("Error list machines: ", err)
	}
	fmt.Println("machines: ", machines)
}
