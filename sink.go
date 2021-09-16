package plogr

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pterm/pterm"
)

// PtermSink implements logr.LogSink.
type PtermSink struct {
	// LevelPrinters maps a pterm.PrefixPrinter to each supported log level.
	LevelPrinters map[int]pterm.PrefixPrinter
	// ErrorPrinter is the instance that formats and styles error messages.
	ErrorPrinter pterm.PrefixPrinter

	keyValues        map[string]interface{}
	messageFormatter func(msg string, keysAndValues map[string]interface{}) string
	scope            string
	writer           io.Writer
}

// ScopeSeparator delimits logger names.
var ScopeSeparator = ":"

// DefaultLevelPrinters contains the default pterm.PrefixPrinter for a specific log levels.
var DefaultLevelPrinters = map[int]pterm.PrefixPrinter{
	0: *pterm.Info.WithPrefix(pterm.Prefix{Text: " INFO  ", Style: pterm.Info.Prefix.Style}),
	1: pterm.Debug,
}

// DefaultFormatter returns a string that looks as following (with colored key/values):
//  * message
//  * message (key="value" foo="bar")
var DefaultFormatter = func(msg string, keysAndValues map[string]interface{}) string {
	if len(keysAndValues) <= 0 {
		return msg
	}
	pairs := make([]string, 0)
	for k, v := range keysAndValues {
		pairs = append(pairs, fmt.Sprintf("%s=\"%+v\"", k, v))
	}
	msg = fmt.Sprintf("%s %s", msg, pterm.FgGray.Sprint("(", strings.Join(pairs, " "), ")"))
	return msg
}

// NewPtermSink returns a new logr.LogSink instance where messages are being printed with pterm.PrefixPrinter.
// PtermSink.LevelPrinters and PtermSink.ErrorPrinter are initialized with DefaultLevelPrinters resp. pterm.Error.
// PtermSink.KeyValueColor is the color that is applied when logging keys and values and defaults to pterm.FgGray.
func NewPtermSink() PtermSink {
	return PtermSink{
		LevelPrinters:    DefaultLevelPrinters,
		ErrorPrinter:     pterm.Error,
		keyValues:        map[string]interface{}{},
		messageFormatter: DefaultFormatter,
		writer:           os.Stdout,
	}
}

// Init implements logr.LogSink.
func (s PtermSink) Init(_ logr.RuntimeInfo) {
	pterm.EnableDebugMessages()
}

// Enabled implements logr.LogSink.
// It will only return true if the PtermSink.LevelPrinters contains the same key as given level.
func (s PtermSink) Enabled(level int) bool {
	_, exists := s.LevelPrinters[level]
	return exists
}

// Info implements logr.LogSink.
func (s PtermSink) Info(level int, msg string, kvs ...interface{}) {
	printer := s.LevelPrinters[level]
	s.print(printer, kvs, msg)
}

// Error implements logr.LogSink.
// The given err is appended to the keys and values array with the "error" key.
func (s PtermSink) Error(err error, msg string, kvs ...interface{}) {
	kvs = append(kvs, "error", err)
	s.print(s.ErrorPrinter, kvs, msg)
}

func (s PtermSink) print(printer pterm.PrefixPrinter, kvs []interface{}, msg string) {
	kvMap := s.toMap(kvs...)
	formatted := s.messageFormatter(msg, kvMap)
	_, _ = fmt.Fprint(s.writer, printer.Sprintln(formatted))
}

// WithValues implements logr.LogSink.
// It returns a new logr.Logger instance that is pre-configured with given keys and values.
func (s PtermSink) WithValues(kvs ...interface{}) logr.LogSink {
	newMap := make(map[string]interface{}, len(s.keyValues)+len(kvs)/2)
	for k, v := range s.keyValues {
		newMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		newMap[kvs[i].(string)] = kvs[i+1]
	}
	return &PtermSink{
		scope:            s.scope,
		keyValues:        newMap,
		LevelPrinters:    s.LevelPrinters,
		ErrorPrinter:     s.ErrorPrinter,
		messageFormatter: s.messageFormatter,
		writer:           s.writer,
	}
}

// WithName implements logr.LogSink.
// It returns a new logr.Logger instance that copies the pterm.PrefixPrinter from previous instance, but modifies the Scope property of the prefix printer.
// The value of the name is joined with the existing name, delimited by ScopeSeparator.
func (s PtermSink) WithName(name string) logr.LogSink {
	newSink := &PtermSink{
		scope:            s.joinName(s.scope, name),
		keyValues:        s.keyValues,
		LevelPrinters:    map[int]pterm.PrefixPrinter{},
		ErrorPrinter:     s.ErrorPrinter,
		messageFormatter: s.messageFormatter,
		writer:           s.writer,
	}
	for level, printer := range s.LevelPrinters {
		newPrinter := printer.WithScope(pterm.Scope{Text: name, Style: printer.Scope.Style})
		newSink.LevelPrinters[level] = *newPrinter
	}
	newSink.ErrorPrinter.Scope.Text = newSink.scope
	return newSink
}

// SetOutput sets the new output on the given sink.
// The difference to WithOutput is that setting the output on this instance also affects other log sinks that were created on the current state of s.
func (s *PtermSink) SetOutput(output io.Writer) *PtermSink {
	s.writer = output
	return s
}

// WithOutput returns a new sink that writes log messages to the given output.
// The difference to SetOutput is that this method doesn't alter the existing sink.
func (s PtermSink) WithOutput(output io.Writer) *PtermSink {
	newSink := &PtermSink{
		scope:            s.scope,
		keyValues:        s.keyValues,
		LevelPrinters:    s.LevelPrinters,
		ErrorPrinter:     s.ErrorPrinter,
		messageFormatter: s.messageFormatter,
		writer:           output,
	}
	return newSink
}

func (s *PtermSink) toMap(kvs ...interface{}) map[string]interface{} {
	if len(kvs)%2 == 1 {
		// Ensure an odd number of items here does not corrupt the list
		kvs = append(kvs, nil)
	}
	kvMap := map[string]interface{}{}

	for k, v := range s.keyValues {
		kvMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i].(string)
		kvMap[key] = kvs[i+1]
	}
	return kvMap
}

func (s *PtermSink) joinName(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return strings.Join([]string{s1, s2}, ScopeSeparator)
}
