package settings

import (
	"path"
	"reflect"
	"strings"
	"testing"
)

func load(file string) (settings *Settings, err error) {
	return Load(path.Join("../../tests/configs", file))
}

func TestNotAFile(t *testing.T) {
	_, err := load("not-a-file.yaml")
	if err == nil {
		t.Error("should fail if no file", err)
	}
}

func TestConfigFile(t *testing.T) {
	s, err := load("config-ok.yaml")
	if err != nil {
		t.Error("could not load settings file", err)
		return
	}
	expected := &Settings{
		Logging: Logging{File: "foo.log", Level: "debug"},
		Name:    "testing-fetter",
		Mode:    "scanner",
		Audit:   Audit{Mode: "reuse"},
		Rules: map[string]Rule{
			"r1": {Paths: []string{"/usr/bin/make"}, Action: "execute", Group: "g1"},
			"r2": {Paths: []string{"/usr/bin/make2"}, Action: "read", Group: "g2", Trigger: "t2"},
			"r3": {Paths: []string{"/root/danger"}, Action: "execute", Trigger: "KILL"},
		},
		Groups: map[string]Group{
			"g1": {RAM: 100, CPU: 10, Pids: 1, Freeze: false},
			"g2": {RAM: 200, CPU: 20, Pids: 0, Freeze: true},
		},
		Triggers: map[string]Trigger{
			"t1": {Run: "/bin/true", Args: []string{"foo", "bar"}, User: "nobody"},
			"t2": {Run: "/bin/false"},
		},
	}
	if !reflect.DeepEqual(s, expected) {
		t.Error("unexpected settings content", s, "vs", expected)
	}
	if s.GetGroup("r1") != "g1" {
		t.Error("bad group for rule")
	}
	if s.GetTrigger("r2") != "t2" {
		t.Error("bad trigger for rule")
	}
}

func TestUnsupportedMode(t *testing.T) {
	_, err := load("config-bad-mode.yaml")
	if err == nil {
		t.Error("Loading config should fail")
		return
	}
	expected := "run mode not supported: garbage"
	if !strings.Contains(err.Error(), expected) {
		t.Error("Should complain of invalid mode", err)
	}
}

func TestRequiredSections(t *testing.T) {
	_, err := load("config-no-rules.yaml")
	if err == nil {
		t.Error("Loading config should fail")
	} else {
		expected := "required key 'rules'"
		if !strings.Contains(err.Error(), expected) {
			t.Error("Should complain of invalid mode", err)
		}
	}
}

func TestRulesMissingTargets(t *testing.T) {
	_, err := load("config-missing-group.yaml")
	if err == nil {
		t.Error("Loading config should fail")
	} else if !strings.Contains(err.Error(), "missing group") {
		t.Error("Should complain of Missing group", err)
	}
	_, err = load("config-missing-trigger.yaml")
	if err == nil {
		t.Error("Loading config should fail")
	} else if !strings.Contains(err.Error(), "missing trigger") {
		t.Error("Should complain of Missing trigger", err)
	}
}
