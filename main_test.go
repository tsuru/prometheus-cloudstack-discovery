package main

import (
	"reflect"
	"testing"

	"github.com/tsuru/prometheus-cloudstack-discovery/cloudstack"
)

func TestMachinesToTg(t *testing.T) {
	machines := []string{"127.0.0.1", "localhost"}
	tgs := machinesToTg(machines, 9090)
	tg := TargetGroup{
		Targets: machines,
		Labels:  map[string]string{"job": "cadvisor"},
	}
	expected := []TargetGroup{tg}
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
