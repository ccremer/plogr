package plogr

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pterm/pterm"
)

// PtermSink implements logr.LogSink.
type PtermSink struct {
	// LevelPrinters maps a pterm.PrefixPrinter to each supported log level.
	LevelPrinters map[int]pterm.PrefixPrinter
	// ErrorPrinter is the instance that formats and styles error messages.
	ErrorPrinter  pterm.PrefixPrinter
	KeyValueColor pterm.Color

	keyValues map[string]interface{}
	scope     string
}

// ScopeSeparator delimits logger names.
var ScopeSeparator = ":"

// KeyJoiner delimits key=value pairs.
var KeyJoiner = " "

// DefaultLevelPrinters contains the default pterm.PrefixPrinter for a specific log levels.
var DefaultLevelPrinters = map[int]pterm.PrefixPrinter{
	0: pterm.Info,
	1: pterm.Debug,
}

// NewPtermSink returns a new logr.LogSink instance where messages are being printed with pterm.PrefixPrinter.
// PtermSink.LevelPrinters and PtermSink.ErrorPrinter are initialized with DefaultLevelPrinters resp. pterm.Error.
// PtermSink.KeyValueColor is the color that is applied when logging keys and values and defaults to pterm.FgGray.
func NewPtermSink() PtermSink {
	return PtermSink{
		LevelPrinters: DefaultLevelPrinters,
		ErrorPrinter:  pterm.Error,
		KeyValueColor: pterm.FgGray,
		keyValues:     map[string]interface{}{},
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
	pairs := s.kvPairs(kvs...)
	if len(pairs) > 0 {
		msg = fmt.Sprintf("%s %s", msg, s.KeyValueColor.Sprint("(", strings.Join(pairs, KeyJoiner), ")"))
	}
	printer.Println(msg)
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
		scope:         s.scope,
		keyValues:     newMap,
		LevelPrinters: s.LevelPrinters,
		ErrorPrinter:  s.ErrorPrinter,
		KeyValueColor: s.KeyValueColor,
	}
}

// WithName implements logr.LogSink.
// It returns a new logr.Logger instance that copies the pterm.PrefixPrinter from previous instance, but modifies the Scope property of the prefix printer.
// The value of the name is joined with the existing name, delimited by ScopeSeparator.
func (s PtermSink) WithName(name string) logr.LogSink {
	newSink := &PtermSink{
		scope:         s.joinName(s.scope, name),
		keyValues:     s.keyValues,
		LevelPrinters: map[int]pterm.PrefixPrinter{},
		ErrorPrinter:  s.ErrorPrinter,
		KeyValueColor: s.KeyValueColor,
	}
	for level, printer := range s.LevelPrinters {
		newPrinter := printer.WithScope(pterm.Scope{Text: name, Style: printer.Scope.Style})
		newSink.LevelPrinters[level] = *newPrinter
	}
	newSink.ErrorPrinter.Scope.Text = newSink.scope
	return newSink
}

func (s *PtermSink) kvPairs(kvs ...interface{}) []string {
	if len(kvs)%2 == 1 {
		// Ensure an odd number of items here does not corrupt the list
		kvs = append(kvs, nil)
	}
	kvPairs := make([]string, 0)

	for k, v := range s.keyValues {
		kvPairs = append(kvPairs, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	for i := 0; i < len(kvs); i += 2 {
		kvPairs = append(kvPairs, fmt.Sprintf("%s=\"%+v\"", kvs[i], kvs[i+1]))
	}
	return kvPairs
}

func (s *PtermSink) joinName(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	return strings.Join([]string{s1, s2}, ScopeSeparator)
}
