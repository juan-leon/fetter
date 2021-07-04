package cgroups

import (
	"reflect"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

func TestEmptySpec(t *testing.T) {
	log.InitLoggerForTests()
	spec := createSpec("foo", &settings.Group{})
	if !reflect.DeepEqual(spec, &specs.LinuxResources{}) {
		t.Error("Should be an empty spec:", spec)
	}
}

func TestFullSpec(t *testing.T) {
	log.InitLoggerForTests()
	spec := createSpec("foo", &settings.Group{CPU: 20, RAM: 4, Pids: 789})
	expected := &specs.LinuxPids{Limit: 789}
	if !reflect.DeepEqual(spec.Pids, expected) {
		t.Error("Pid spec", spec.Pids, "should be", expected)
	}
	if *spec.Memory.Limit != int64(4*1024*1024) {
		t.Error("Bad memory limit")
	}
	if *spec.CPU.Period != uint64(1000000) {
		t.Error("Bad cpu period")
	}
	if *spec.CPU.Quota != int64(200000*numCPUs) {
		t.Error("Bad cpu quota")
	}
}
