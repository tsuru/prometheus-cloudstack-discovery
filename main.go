package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tsuru/prometheus-cloudstack-discovery/cloudstack"
)

// TargetGroup is a collection of related hosts that prometheus monitors
type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

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

func in(v string, list []string) bool {
	for _, i := range list {
		if v == i {
			return true
		}
	}
	return false
}

func filterProjects(projects []cloudstack.Project, ignore []string) []cloudstack.Project {
	var filteredProjects []cloudstack.Project
	for _, p := range projects {
		if !in(p.Id, ignore) {
			filteredProjects = append(filteredProjects, p)
		}
	}
	return filteredProjects
}

func listMachines(c *cloudstack.Client, projectsToIgnore []string) ([]string, error) {
	params := map[string]string{"simple": "true"}
	var response cloudstack.ListProjectsResponse
	err := c.Do("listProjects", params, &response)
	if err != nil {
		return nil, err
	}
	projects := response.ListProjectsResponse.Project
	projects = filterProjects(projects, projectsToIgnore)
	mc := make(chan []string)
	for _, p := range projects {
		go listMachineByProject(c, p.Id, mc)
	}
	var machines []string
	for range projects {
		machines = append(machines, <-mc...)
	}
	close(mc)
	return machines, err
}

func main() {
	log.SetOutput(ioutil.Discard)
	var (
		address        = flag.String("url", "", "cloudstack api url address")
		apiKey         = flag.String("api-key", "", "cloudstack api key")
		secretKey      = flag.String("secret-key", "", "cloudstack secret key")
		sleep          = flag.Duration("sleep", 0, "Amount of time between regenerating the target_group.json")
		dest           = flag.String("dest", "", "File to write the target group JSON. (e.g. `tgroups/target_groups.json`)")
		port           = flag.Int("port", 80, "Port that is exposing /metrics")
		ignoreProjects = flag.String("ignore-projects", "", "List of project ids to be ignored")
		projectId      = flag.String("project-id", "", "Filter by project-id")
	)
	flag.Parse()
	c := &cloudstack.Client{
		ApiKey:    url.QueryEscape(*apiKey),
		SecretKey: url.QueryEscape(*secretKey),
		URL:       *address,
	}
	for {
		machines, err := listMachines(c, strings.Split(*ignoreProjects, ","))
		if err != nil {
			log.Fatal("Error list machines: ", err)
		}
		targetGroups := machinesToTg(machines, *port)
		b, err := json.Marshal(targetGroups)
		if err != nil {
			log.Fatal("Error marshal json: ", err)
		}
		if *dest == "" {
			fmt.Println(string(b))
		} else {
			err = atomicWriteFile(*dest, b, ".new")
		}
		if err != nil {
			log.Fatal(err)
		}
		if *sleep == 0 {
			break
		}
		time.Sleep(*sleep)
	}
}

func machinesToTg(machines []string, port int) []TargetGroup {
	for i := range machines {
		machines[i] = fmt.Sprintf("%s:%d", machines[i], port)
	}
	tg := TargetGroup{
		Targets: machines,
		Labels:  map[string]string{"job": "cadvisor"},
	}
	return []TargetGroup{tg}
}

func atomicWriteFile(filename string, data []byte, tmpSuffix string) error {
	err := ioutil.WriteFile(filename+tmpSuffix, data, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(filename+tmpSuffix, filename)
	if err != nil {
		return err
	}
	return nil
}
