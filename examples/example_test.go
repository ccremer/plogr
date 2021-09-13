// +build examples

package examples

import (
	"errors"
	"testing"

	"github.com/ccremer/plogr"
	"github.com/go-logr/logr"
	"github.com/pterm/pterm"
)

func TestExample_PtermSink_Error(t *testing.T) {
	sink := plogr.NewPtermSink()
	logger := logr.New(sink)
	logger.Error(errors.New("this is an error"), "additional error message", "key", "value")
}

func TestExample_PtermSink_Info(t *testing.T) {
	sink := plogr.NewPtermSink()
	logger := logr.New(sink)
	logger.Info("info message", "key", "value")
}

func TestExample_PtermSink_WithName(t *testing.T) {
	sink := plogr.NewPtermSink()
	logger := logr.New(sink)
	logger.WithName("scope").Info("this message should print with a scope")
	logger.WithName("error").WithName("scope").Error(errors.New("this is an error"), "this message should print with a nested scope")
	logger.Info("this should NOT print a scope")
	logger.Error(errors.New("not an error"), "this should NOT print a scope")
}

func TestExample_PtermSink_WithValues(t *testing.T) {
	sink := plogr.NewPtermSink()
	logger := logr.New(sink)
	logger.WithName("values").WithValues("key", "value").Info("this message should print with values", "foo", "bar")
}

func TestExample_PtermSink_Debug(t *testing.T) {
	sink := plogr.NewPtermSink()
	logger := logr.New(sink)
	logger.V(5).Info("This message does not get printed", "reason", "level doesn't exist", "level", 5)
	logger.V(1).Info("debug message that actually gets printed", "key", "value", "level", 1)
}

func TestExample_PtermSink_MoreLevels(t *testing.T) {
	sink := plogr.NewPtermSink()
	sink.LevelPrinters[0] = pterm.Warning
	sink.LevelPrinters[1] = pterm.Success
	sink.LevelPrinters[2] = pterm.Info
	sink.LevelPrinters[3] = pterm.Debug
	logger := logr.New(sink)
	logger.V(0).Info("Warning message")
	logger.V(1).Info("Success message")
	logger.V(2).Info("Info message")
	logger.V(3).Info("Debug message")
}
