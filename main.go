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

func listMachineByProject(c *cloudstack.Client, projectID string, mc chan []cloudstack.VirtualMachine) {
	var machines []cloudstack.VirtualMachine
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
		machines = append(machines, vm)
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

func listMachines(c *cloudstack.Client, projectIDs []string, projectsToIgnore []string) ([]cloudstack.VirtualMachine, error) {
	params := map[string]string{"simple": "true"}
	var response cloudstack.ListProjectsResponse
	err := c.Do("listProjects", params, &response)
	if err != nil {
		return nil, err
	}
	projects := []cloudstack.Project{}
	if len(projectIDs) > 0 {
		for _, id := range projectIDs {
			projects = append(projects, cloudstack.Project{Id: id})
		}
	} else {
		projects = response.ListProjectsResponse.Project
		projects = filterProjects(projects, projectsToIgnore)
	}
	mc := make(chan []cloudstack.VirtualMachine)
	for _, p := range projects {
		go listMachineByProject(c, p.Id, mc)
	}
	var machines []cloudstack.VirtualMachine
	for range projects {
		machines = append(machines, <-mc...)
	}
	close(mc)
	return machines, err
}

func main() {
	var (
		address        = flag.String("url", "", "cloudstack api url address")
		apiKey         = flag.String("api-key", "", "cloudstack api key")
		secretKey      = flag.String("secret-key", "", "cloudstack secret key")
		sleep          = flag.Duration("sleep", 0, "Amount of time between regenerating the target_group.json")
		dest           = flag.String("dest", "", "File to write the target group JSON. (e.g. `tgroups/target_groups.json`)")
		port           = flag.Int("port", 80, "Port that is exposing /metrics")
		ignoreProjects = flag.String("ignore-projects", "", "List of project ids to be ignored separated by comma")
		projects       = flag.String("projects", "", "Filter by a list of project-id separared by comma")
		job            = flag.String("job", "cadvisor", "Prometheus job name to label targets")
		tagName        = flag.String("tag", "", "Cloudstack VM Tag with job/port list. (e.g, `PROMETHEUS_ENDPOINTS` where PROMETHEUS_ENDPOINTS=cadvisor/9094,node-exporter/9095)")
	)
	flag.Parse()
	c := &cloudstack.Client{
		ApiKey:    url.QueryEscape(*apiKey),
		SecretKey: url.QueryEscape(*secretKey),
		URL:       *address,
	}
	p := []string{}
	if *projects != "" {
		p = strings.Split(*projects, ",")
	}
	ip := []string{}
	if *ignoreProjects != "" {
		ip = strings.Split(*ignoreProjects, ",")
	}
	for {
		machines, err := listMachines(c, p, ip)
		if err != nil {
			log.Fatal("Error list machines: ", err)
		}
		targetGroups := machinesToTg(machines, *port, *job, *tagName)
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

func machinesToTg(machines []cloudstack.VirtualMachine, port int, job string, tagName string) []TargetGroup {
	var targetGroups []TargetGroup
	for _, m := range machines {
		targetGroups = append(targetGroups, targetsFromTag(m, tagName)...)
		if job != "" && port != 0 {
			var targets []string
			for _, n := range m.Nic {
				targets = append(targets, fmt.Sprintf("%s:%d", n.IpAddress, port))
			}
			targetGroups = append(targetGroups, TargetGroup{
				Targets: targets,
				Labels:  map[string]string{"job": job, "project": m.Project, "displayname": m.Displayname},
			})
		}
	}
	return targetGroups
}

func targetsFromTag(m cloudstack.VirtualMachine, tagName string) []TargetGroup {
	if tagName == "" {
		return nil
	}
	var targetGroups []TargetGroup
	for _, t := range m.Tags {
		if tagName != t.Key {
			continue
		}
		tagValues := strings.Split(t.Value, ",")
		for _, v := range tagValues {
			tagJob, tagPort := splitJobPort(v)
			var targets []string
			for _, n := range m.Nic {
				targets = append(targets, fmt.Sprintf("%s:%s", n.IpAddress, tagPort))
			}
			targetGroups = append(targetGroups, TargetGroup{
				Targets: targets,
				Labels:  map[string]string{"job": tagJob, "project": m.Project, "displayname": m.Displayname},
			})
		}
	}
	return targetGroups
}

func splitJobPort(s string) (string, string) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
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
