package scanner

import (
	"os/exec"
	"testing"
	"time"

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
	config := &settings.Settings{
		Rules: map[string]settings.Rule{
			"r1": {Paths: []string{"/bin/sleep"}, Action: "execute", Group: "g1"},
			"r2": {Paths: []string{"/bin/sleep"}, Action: "execute", Group: "g1"},
		},
	}
	mock := fakeMover{}
	ps := NewProcessScanner(config, &mock)
	exec.Command("/bin/sleep", "2").Start()
	time.Sleep(time.Millisecond * 100)
	ps.Scan()
	if mock.pid == 0 {
		t.Error("We should have detected the pid")
	}
	if mock.where != "g1" {
		t.Error("We should have move the process into group 'g1'")
	}
}
