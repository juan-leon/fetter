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
		t.Error("shoul fail if no file", err)
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
		Rules: []Rule{
			{Path: "/usr/bin/make", Action: "execute", Group: "compilation"},
			{Path: "/usr/bin/make2", Action: "read", Group: "compilation"},
		},
		Groups: []Group{
			{Name: "g1", RAM: 100, CPU: 10, Pids: 1, Freeze: false},
			{Name: "g2", RAM: 200, CPU: 20, Pids: 0, Freeze: true},
		},
	}
	if !reflect.DeepEqual(s, expected) {
		t.Error("unexpected override names", s, expected)
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
	_, err := load("config-no-groups.yaml")
	if err == nil {
		t.Error("Loading config should fail")
	} else {
		expected := "required key 'groups'"
		if !strings.Contains(err.Error(), expected) {
			t.Error("Should complain of invalid mode", err)
		}
	}
	_, err = load("config-no-rules.yaml")
	if err == nil {
		t.Error("Loading config should fail")
	} else {
		expected := "required key 'rules'"
		if !strings.Contains(err.Error(), expected) {
			t.Error("Should complain of invalid mode", err)
		}
	}
}
