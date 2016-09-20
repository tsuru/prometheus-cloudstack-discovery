package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/atsaki/golang-cloudstack-library"
)

func main() {
	log.SetOutput(ioutil.Discard)
	var (
		address   = flag.String("url", "", "cloudstack api url address")
		apiKey    = flag.String("api-key", "", "cloudstack api key")
		secretKey = flag.String("secret-key", "", "cloudstack secret key")
	)
	flag.Parse()
	endpoint, err := url.Parse(*address)
	if err != nil {
		log.Fatal("Error parsing url:", err)
	}
	client, err := cloudstack.NewClient(
		endpoint,
		url.QueryEscape(*apiKey),
		url.QueryEscape(*secretKey),
		"",
		"",
	)
	if err != nil {
		log.Fatal("Error creating client: ", err)
	}
	fmt.Println("client: ", client)
	projectsParams := cloudstack.NewListProjectsParameter()
	projects, err := client.ListProjects(projectsParams)
	if err != nil {
		log.Fatal("Error list projects: ", err)
	}
	fmt.Println("projects: ", projects)
	var machines []*cloudstack.VirtualMachine
	for _, p := range projects {
		machinesParams := cloudstack.NewListVirtualMachinesParameter()
		machinesParams.ProjectId = p.Id
		m, err := client.ListVirtualMachines(machinesParams)
		if err != nil {
			fmt.Printf("Error list machines for project %s: %s\n", p.Id, err)
		} else {
			machines = append(machines, m...)
		}
	}
	fmt.Println("machines: ", machines)
}
