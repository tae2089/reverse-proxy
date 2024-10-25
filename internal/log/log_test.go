package log

import "testing"

func TestInfo(t *testing.T) {
	Info("This is an info message")
}

func TestError(t *testing.T) {
	Error("This is an error message")
}

func TestDebug(t *testing.T) {
	Debug("This is a debug message")
}

func TestWarn(t *testing.T) {
	Warn("This is a warning message")
}
