package triggers

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

var triggers = map[string]settings.Trigger{
	"t1": {Run: "/bin/true", Args: []string{"foo", "bar"}},
	"t2": {Run: "/bin/false", Args: []string{"foo", "bar"}, User: "nobody"},
}
var config = &settings.Settings{Triggers: triggers}

func TestNoTrigger(t *testing.T) {
	tr := NewTriggerRunner(config)
	err := tr.Run("No-trigger", nil)
	if err == nil {
		t.Error("no trigger present should return an error")
	}
}

func TestTriggerTrue(t *testing.T) {
	tr := NewTriggerRunner(config)
	err := tr.Run("t1", nil)
	if err != nil {
		t.Error("trigger should not return an error", err)
	}
}

func TestRunTrue(t *testing.T) {
	log.InitLoggerForTests()
	err := run(&settings.Trigger{Run: "/bin/true"}, "true", nil)
	if err != nil {
		t.Error("We could not even run '/bin/true'", err)
	}
}

func TestRunWithBadUser(t *testing.T) {
	log.InitLoggerForTests()
	err := run(&settings.Trigger{Run: "/bin/true", User: "/dev/null"}, "true", nil)
	if err == nil {
		t.Error("We should have triggered an error")
	}
	if !strings.Contains(err.Error(), "unknown user") {
		t.Error("We should have complained about unknown user")
	}
}

func TestRunTool(t *testing.T) {
	log.InitLoggerForTests()
	token := "token33"
	file := ".fetter.test"
	path := "/tmp/.fetter.test"
	value := "magic-val-22"
	err := run(
		&settings.Trigger{
			Run:  "../../tests/scripts/trigger.sh",
			Args: []string{token, file},
			User: "root",
		},
		"foo",
		&map[string]string{"Var1": value, "VAR2": "value2"},
	)
	if err != nil {
		t.Error("Tool failed to execute:", err)
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error("Tool failed to generate file:", err)
	}
	s := string(bytes)
	if !strings.Contains(s, token) {
		t.Error("Token is not in generated file")
	}
	if !strings.Contains(s, value) {
		t.Error("Variable value is not in generated file")
	}
	os.Remove(path)
}

func TestRunFailingTool(t *testing.T) {
	log.InitLoggerForTests()
	err := run(
		&settings.Trigger{
			Run:  "/bin/ls",
			Args: []string{"/not/a/file"},
		},
		"foo",
		nil,
	)
	if err == nil {
		t.Error("Tool should report an error")
	}
}
