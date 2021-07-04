package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

type fakeMover struct {
	pid   int
	where string
}

func (f *fakeMover) Move(pid int, cgroup string) error {
	f.pid = pid
	f.where = cgroup
	return nil
}

func TestScan(t *testing.T) {
	log.InitLoggerForTests()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal("Test cannot continue; failed to find command", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		t.Fatal("Test cannot continue; failed to resolve symlinks", executable, err)
	}
	if _, err := os.Stat(executable); err != nil {
		t.Fatal("Test cannot continue; failed to find executable", executable, err)
	}
	config := &settings.Settings{
		Rules: map[string]settings.Rule{
			"r1": {Paths: []string{executable}, Action: "execute", Group: "g1"},
			"r2": {Paths: []string{executable}, Action: "execute", Group: "g1"},
		},
	}
	mock := fakeMover{}
	ps := NewProcessScanner(config, &mock)
	ps.Scan()
	if mock.pid == 0 {
		t.Error("We should have detected the pid")
	}
	if mock.where != "g1" {
		t.Error("We should have move the process into group 'g1'")
	}
}
