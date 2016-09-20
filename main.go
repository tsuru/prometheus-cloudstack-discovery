package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/tsuru/tsuru/net"
)

type client struct {
	apiKey    string
	secretKey string
	url       string
}

type ListVirtualMachinesResponse struct {
	ListVirtualMachinesResponse struct {
		VirtualMachine []VirtualMachine `json:"virtualmachine"`
	} `json:"listvirtualmachinesresponse"`
}

type VirtualMachine struct {
	Nic []NicStruct `json:"nic"`
}

type NicStruct struct {
	IpAddress string `json:"ipaddress"`
}

type ListProjectsResponse struct {
	ListProjectsResponse struct {
		Project []Project `json:"project"`
	} `json:"listprojectsresponse"`
}

type Project struct {
	Id string
}

func listMachineByProject(c *client, projectID string, mc chan []string) {
	params := map[string]string{
		"projectid": projectID,
		"simple":    "true",
	}
	var m ListVirtualMachinesResponse
	c.do("listVirtualMachines", params, &m)
	var machines []string
	for _, vm := range m.ListVirtualMachinesResponse.VirtualMachine {
		for _, n := range vm.Nic {
			machines = append(machines, n.IpAddress)
		}
	}
	mc <- machines
}

func listMachines(c *client) ([]string, error) {
	params := map[string]string{"simple": "true"}
	var projects ListProjectsResponse
	err := c.do("listProjects", params, &projects)
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
	c := &client{
		apiKey:    url.QueryEscape(*apiKey),
		secretKey: url.QueryEscape(*secretKey),
		url:       *address,
	}
	fmt.Println(c)
	machines, err := listMachines(c)
	if err != nil {
		log.Fatal("Error list machines: ", err)
	}
	fmt.Println("machines: ", machines)
}

func (c *client) buildUrl(command string, params map[string]string) (string, error) {
	params["command"] = command
	params["response"] = "json"
	params["apiKey"] = c.apiKey
	var sorted_keys []string
	for k := range params {
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	var string_params []string
	for _, key := range sorted_keys {
		queryStringParam := fmt.Sprintf("%s=%s", key, url.QueryEscape(params[key]))
		string_params = append(string_params, queryStringParam)
	}
	queryString := strings.Join(string_params, "&")
	digest := hmac.New(sha1.New, []byte(c.secretKey))
	digest.Write([]byte(strings.ToLower(queryString)))
	signature := base64.StdEncoding.EncodeToString(digest.Sum(nil))
	return fmt.Sprintf("%s?%s&signature=%s", c.url, queryString, url.QueryEscape(signature)), nil
}

func (c *client) do(cmd string, params map[string]string, result interface{}) error {
	u, err := c.buildUrl(cmd, params)
	if err != nil {
		return err
	}
	client := net.Dial5Full300ClientNoKeepAlive
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code for %s command %d: %s", cmd, resp.StatusCode, string(body))
	}
	if result != nil {
		err = json.Unmarshal(body, result)
		if err != nil {
			return fmt.Errorf("Unexpected result data for %s command: %s - Body: %s", cmd, err.Error(), string(body))
		}
	}
	return nil
}
