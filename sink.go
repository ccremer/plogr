package plogr

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pterm/pterm"
)

// PtermSink implements logr.LogSink.
type PtermSink struct {
	// LevelPrinters maps a pterm.PrefixPrinter to each supported log level.
	LevelPrinters map[int]pterm.PrefixPrinter
	// LevelEnabled enables or disables logging for the supported log levels.
	LevelEnabled map[int]bool

	// ErrorPrinter is the instance that formats and styles error messages.
	ErrorPrinter pterm.PrefixPrinter
	// FallbackPrinter is the instance that is used for a level that doesn't exist in LevelPrinters.
	// If nil, no fallback is used and the log message gets discarded.
	FallbackPrinter *pterm.PrefixPrinter

	keyValues        map[string]interface{}
	messageFormatter func(msg string, keysAndValues map[string]interface{}) string
	scope            string
}

// ScopeSeparator delimits logger names.
var ScopeSeparator = ":"

// DefaultLevelPrinters contains the default pterm.PrefixPrinter for a specific log levels.
var DefaultLevelPrinters = map[int]pterm.PrefixPrinter{
	0: *pterm.Info.WithPrefix(pterm.Prefix{Text: " INFO  ", Style: pterm.Info.Prefix.Style}),
	1: pterm.Debug,
}

// DefaultErrorPrinter is the default pterm.PrefixPrinter for the error level.
var DefaultErrorPrinter = *pterm.Error.WithShowLineNumber(true).WithLineNumberOffset(2)

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
		LevelPrinters: DefaultLevelPrinters,
		LevelEnabled: map[int]bool{
			0: true,
			1: true,
		},
		ErrorPrinter:     DefaultErrorPrinter,
		keyValues:        map[string]interface{}{},
		messageFormatter: DefaultFormatter,
	}
}

// Init implements logr.LogSink.
func (s PtermSink) Init(_ logr.RuntimeInfo) {
	pterm.EnableDebugMessages()
}

// Enabled implements logr.LogSink.
// It will return false
//  * if LevelEnabled has a key with the level and a value "false"
//  * if LevelPrinters does not contain the requested level as key and FallbackPrinter is nil
func (s PtermSink) Enabled(level int) bool {
	_, exists := s.LevelPrinters[level]
	if exists {
		enabled, defined := s.LevelEnabled[level]
		if defined {
			return enabled
		}
		return true
	}
	return s.FallbackPrinter != nil
}

// Info implements logr.LogSink.
func (s PtermSink) Info(level int, msg string, kvs ...interface{}) {
	printer, found := s.LevelPrinters[level]
	if !found {
		// even though FallbackPrinter may be nil, it should be safe to use since Enabled() guards calling Info if fallback is disabled.
		printer = *s.FallbackPrinter
	}
	s.print(printer, kvs, msg)
}

// Error implements logr.LogSink.
// The given err is appended to the keys and values array with the "error" key, but only if err is non-nil.
func (s PtermSink) Error(err error, msg string, kvs ...interface{}) {
	if err != nil {
		kvs = append(kvs, "error", err)
	}
	s.print(s.ErrorPrinter, kvs, msg)
}

func (s PtermSink) print(printer pterm.PrefixPrinter, kvs []interface{}, msg string) {
	kvMap := s.toMap(kvs...)
	formatted := s.messageFormatter(msg, kvMap)
	printer.Printfln(formatted)
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
	cpy := s.copy()
	cpy.keyValues = newMap
	return &cpy
}

// WithName implements logr.LogSink.
// It returns a new logr.Logger instance that copies the pterm.PrefixPrinter from previous instance, but modifies the Scope property of the prefix printer.
// The value of the name is joined with the existing name, delimited by ScopeSeparator.
func (s PtermSink) WithName(name string) logr.LogSink {
	newSink := s.copy()
	newSink.scope = s.joinName(s.scope, name)
	newSink.LevelPrinters = map[int]pterm.PrefixPrinter{}
	for level, printer := range s.LevelPrinters {
		newPrinter := printer.WithScope(pterm.Scope{Text: name, Style: printer.Scope.Style})
		newSink.LevelPrinters[level] = *newPrinter
	}
	newSink.ErrorPrinter.Scope.Text = newSink.scope
	if newSink.FallbackPrinter != nil {
		newSink.FallbackPrinter.Scope.Text = newSink.scope
	}
	return newSink
}

// SetOutput is a convenience func that sets the new writer directly in pterm.PrefixPrinter.
// The difference to WithOutput is that setting the output on this instance also affects other log sinks that were created on the current state of s.
func (s *PtermSink) SetOutput(output io.Writer) *PtermSink {
	for i, printer := range s.LevelPrinters {
		s.LevelPrinters[i] = *printer.WithWriter(output)
	}
	return s
}

// WithOutput returns a new sink that writes log messages to the given output.
// The difference to SetOutput is that this method doesn't alter the existing sink.
func (s PtermSink) WithOutput(output io.Writer) *PtermSink {
	newSink := s.copy()
	newSink.LevelPrinters = map[int]pterm.PrefixPrinter{}
	for i, printer := range s.LevelPrinters {
		newSink.LevelPrinters[i] = *printer.WithWriter(output)
	}
	return &newSink
}

// Name returns the currently configured scope name
func (s PtermSink) Name() string {
	return s.scope
}

// SetLevelEnabled explicitly enables or disables a logging level.
func (s *PtermSink) SetLevelEnabled(level int, enabled bool) *PtermSink {
	s.LevelEnabled[level] = enabled
	return s
}

// WithFallbackPrinter returns a copy of s with the given fallback printer.
// The fallback printer is used for levels that aren't explicitly defined in LevelPrinters.
// pterm.Debug is a good starting point for a fallback printer.
func (s PtermSink) WithFallbackPrinter(printer pterm.PrefixPrinter) *PtermSink {
	cpy := s.copy()
	cpy.FallbackPrinter = &printer
	return &cpy
}

func (s PtermSink) copy() PtermSink {
	return PtermSink{
		LevelPrinters:    s.LevelPrinters,
		LevelEnabled:     s.LevelEnabled,
		ErrorPrinter:     s.ErrorPrinter,
		FallbackPrinter:  s.FallbackPrinter,
		keyValues:        s.keyValues,
		messageFormatter: s.messageFormatter,
		scope:            s.scope,
	}
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
