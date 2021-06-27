package log

import (
	"testing"

	"github.com/juan-leon/fetter/pkg/settings"
)

func TestInitLog(t *testing.T) {
	InitConsoleLogger()
	Console.Sync()
	InitFileLogger(settings.Logging{File: "/dev/null", Level: "none"})
	InitLoggerForTests()
}
