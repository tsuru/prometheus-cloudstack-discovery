package main

import (
	"reflect"
	"testing"
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
