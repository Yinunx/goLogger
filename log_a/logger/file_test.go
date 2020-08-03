package logger

import (
	"testing"
)

/*
func TestFileLogger(t *testing.T) {
	logger := NewFileLogger(LogLevelDebug, "c:/loggos/", "test")
	logger.Debug("user id is[%d] is com from china", 32445)
	logger.Warn("test warn log")
	logger.Fatal("test fatal log")
	logger.Close()
}
*/

func TestConsoleLogger(t *testing.T) {
	logger := NewConsoleLogger(LogLevelDebug)
	logger.Debug("user id is[%d] is com from china", 32445)
	logger.Warn("test warn log")
	logger.Fatal("test fatal log")
	logger.Close()
}
