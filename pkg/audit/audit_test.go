package audit

import (
	"testing"

	"github.com/juan-leon/fetter/pkg/settings"
)

func TestAuditModeAssertion(t *testing.T) {
	if !assertAuditMode(MODE_PRESERVE) {
		t.Error(MODE_PRESERVE, "is a valid mode")
	}
	if assertAuditMode("foobar") {
		t.Error("foobar is not a valid mode")
	}
}

func TestValidateRule(t *testing.T) {
	if validateRule(settings.Rule{Path: "foo", Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Action: "foo", Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Path: "foo", Action: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Path: "foo", Action: "foo", Group: "foo"}) == nil {
		t.Error("Rule should fail validation")
	}
	if validateRule(settings.Rule{Path: "foo", Action: SYSCALL_EXECUTE, Group: "foo"}) != nil {
		t.Error("Rule should pass validation")
	}
}

func TestRuleFormat(t *testing.T) {
	value := asAuditFmt(settings.Rule{Path: "/foo", Group: "danger", Action: SYSCALL_EXECUTE})
	expected := "-w /foo -p x -k cgroup_danger"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
	value = asAuditFmt(settings.Rule{Path: "foo", Group: "foobar", Action: SYSCALL_READ})
	expected = "-w foo -p r -k cgroup_foobar"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
	value = asAuditFmt(settings.Rule{Path: "none", Group: "test", Action: SYSCALL_WRITE})
	expected = "-w none -p w -k cgroup_test"
	if value != expected {
		t.Error("Rule should be formatted as", expected, "instead of", value)
	}
}
