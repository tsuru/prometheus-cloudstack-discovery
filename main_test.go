package main

import (
	"reflect"
	"testing"

	"github.com/tsuru/prometheus-cloudstack-discovery/cloudstack"
)

func TestMachinesToTg(t *testing.T) {
	machines := []cloudstack.VirtualMachine{{
		Project:     "project_name",
		Displayname: "display_name",
		Nic:         []cloudstack.NicStruct{{IpAddress: "127.0.0.1"}},
	}, {
		Project:     "project_name2",
		Displayname: "display_name2",
		Nic:         []cloudstack.NicStruct{{IpAddress: "localhost"}},
	}, {
		Project:     "project_name3",
		Displayname: "display_name3",
		Nic:         []cloudstack.NicStruct{{IpAddress: "localhost"}},
		Tags:        []cloudstack.Tag{{Key: "abc", Value: "1"}, {Key: "PROMETHEUS_ENDPOINTS", Value: "node-exporter/9095,tsuru/8080"}},
	}}
	tgs := machinesToTg(machines, []string{"cadvisor/9090"}, "PROMETHEUS_ENDPOINTS")
	expected := []TargetGroup{{
		Targets: []string{"127.0.0.1:9090"},
		Labels:  map[string]string{"job": "cadvisor", "project": "project_name", "displayname": "display_name"},
	}, {
		Targets: []string{"localhost:9090"},
		Labels:  map[string]string{"job": "cadvisor", "project": "project_name2", "displayname": "display_name2"},
	}, {
		Targets: []string{"localhost:9095"},
		Labels:  map[string]string{"job": "node-exporter", "project": "project_name3", "displayname": "display_name3"},
	}, {
		Targets: []string{"localhost:8080"},
		Labels:  map[string]string{"job": "tsuru", "project": "project_name3", "displayname": "display_name3"},
	}, {
		Targets: []string{"localhost:9090"},
		Labels:  map[string]string{"job": "cadvisor", "project": "project_name3", "displayname": "display_name3"},
	}}
	if !reflect.DeepEqual(tgs, expected) {
		t.Errorf("machinesToTg(%q) == %q, want %q", machines, tgs, expected)
	}
}

func TestFilterProjects(t *testing.T) {
	projects := []cloudstack.Project{{Id: "1"}, {Id: "2"}, {Id: "3"}, {Id: "4"}}
	projectsToIgnore := []string{"3", "4"}
	projects = filterProjects(projects, projectsToIgnore)
	expected := []cloudstack.Project{{Id: "1"}, {Id: "2"}}
	if !reflect.DeepEqual(projects, expected) {
		t.Errorf("projects are %q, want %q", projects, expected)
	}
}
