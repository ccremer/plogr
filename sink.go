package plogr

import (
	"github.com/go-logr/logr"
)

type PtermSink struct {
}

func (PtermSink) Init(info logr.RuntimeInfo) {
	panic("implement me")
}

func (PtermSink) Enabled(level int) bool {
	panic("implement me")
}

func (PtermSink) Info(level int, msg string, keysAndValues ...interface{}) {
	panic("implement me")
}

func (PtermSink) Error(err error, msg string, keysAndValues ...interface{}) {
	panic("implement me")
}

func (PtermSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	panic("implement me")
}

func (PtermSink) WithName(name string) logr.LogSink {
	panic("implement me")
}
