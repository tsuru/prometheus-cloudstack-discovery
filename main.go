package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/atsaki/golang-cloudstack-library"
)

func main() {
	var (
		address   = flag.String("url", "", "cloudstack api url address")
		apiKey    = flag.String("api-key", "", "cloudstack api key")
		secretKey = flag.String("secret-key", "", "cloudstack secret key")
	)
	endpoint, err := url.Parse(*address)
	if err != nil {
		log.Fatal(err)
	}
	client, err := cloudstack.NewClient(endpoint, *apiKey, *secretKey, "", "")
	if err != nil {
		log.Fatal(err)
	}
	params := cloudstack.NewListVirtualMachinesParameter()
	machines, err := client.ListVirtualMachines(params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(machines)
}
