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
	}}
	tgs := machinesToTg(machines, 9090, "cadvisor")
	expected := []TargetGroup{{
		Targets: []string{"127.0.0.1:9090"},
		Labels:  map[string]string{"job": "cadvisor", "project": "project_name", "displayname": "display_name"},
	}, {
		Targets: []string{"localhost:9090"},
		Labels:  map[string]string{"job": "cadvisor", "project": "project_name2", "displayname": "display_name2"},
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
