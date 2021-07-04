package audit

import (
	"testing"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

var rules = map[string]settings.Rule{
	"r1": {Paths: []string{"none"}, Action: "execute", Trigger: "t1"},
	"r2": {Paths: []string{"none"}, Action: "read", Group: "g1"},
}
var config = &settings.Settings{
	Rules: rules,
	Mode:  "bad-mode",
}

type mock struct {
	moved bool
	ran   bool
}

func (m *mock) Move(pid int, cgroup string) error {
	m.moved = true
	return nil
}

func (m *mock) Run(name string, data *map[string]string) error {
	m.ran = true
	return nil
}

func TestAuditModeAssertion(t *testing.T) {
	if !assertAuditMode(MODE_PRESERVE) {
		t.Error(MODE_PRESERVE, "is a valid mode")
	}
	if assertAuditMode("foobar") {
		t.Error("foobar is not a valid mode")
	}
}

func TestValidateRule(t *testing.T) {
	if validateRule(settings.Rule{Paths: []string{"foo"}, Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Action: "foo", Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Paths: []string{"foo"}, Action: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Paths: []string{"foo"}, Action: "foo", Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Paths: []string{"foo"}, Action: SYSCALL_EXECUTE, Group: "foo"}) != nil {
		t.Error("Rule should pass validation")
	}
}

func TestRuleFormat(t *testing.T) {
	value := asAuditFmt("danger", "/foo", SYSCALL_EXECUTE)
	expected := "-w /foo -p x -k fetter_danger"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
	value = asAuditFmt("foobar", "foo", SYSCALL_READ)
	expected = "-w foo -p r -k fetter_foobar"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
	value = asAuditFmt("test", "none", SYSCALL_WRITE)
	expected = "-w none -p w -k fetter_test"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
}

func TestFakeRule(t *testing.T) {
	log.InitLoggerForTests()
	m := &mock{}
	scl := SysCallListener{
		client:     nil,
		config:     config,
		procMover:  m,
		procRunner: m,
	}
	scl.processMatch(1, "fake-rule", nil)
	if m.moved {
		t.Error("No process should be moved here")
	}
	if m.ran {
		t.Error("No process should be triggered here")
	}
}

func TestTrigger(t *testing.T) {
	log.InitLoggerForTests()
	m := &mock{}
	scl := SysCallListener{
		client:     nil,
		config:     config,
		procMover:  m,
		procRunner: m,
	}
	scl.processMatch(1, "r1", nil)
	if m.moved {
		t.Error("No process should be moved here")
	}
	if !m.ran {
		t.Error("Process should have been triggered here")
	}
}

func TestMove(t *testing.T) {
	log.InitLoggerForTests()
	m := &mock{}
	scl := SysCallListener{
		client:     nil,
		config:     config,
		procMover:  m,
		procRunner: m,
	}
	scl.processMatch(1, "r2", nil)
	if !m.moved {
		t.Error("process should have been moved")
	}
	if m.ran {
		t.Error("No process should be triggered here")
	}
}

func TestBuildWithBadModeShouldReturnNil(t *testing.T) {
	log.InitLoggerForTests()
	m := &mock{}
	scl := NewSysCallListener(config, m, m)
	if scl != nil {
		t.Error("No SysCallListener should be created here")
	}
}
