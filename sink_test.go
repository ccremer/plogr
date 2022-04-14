package plogr

import (
	"bytes"
	"testing"

	"github.com/go-logr/logr"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPtermSink_DefaultFormatter(t *testing.T) {
	tests := map[string]struct {
		givenKeysAndValues map[string]interface{}
		expectedResult     string
	}{
		"GivenNil_ThenReturnEmptyString": {
			givenKeysAndValues: nil,
			expectedResult:     "message",
		},
		"GivenEmptyList_ThenReturnEmptyString": {
			givenKeysAndValues: map[string]interface{}{},
			expectedResult:     "message",
		},
		"GivenSingleEntry_WhenValueNil_ThenReturnNilRepresentation": {
			givenKeysAndValues: map[string]interface{}{"key": nil},
			expectedResult:     "message \x1b[90m(key=\"<nil>\")\x1b[0m",
		},
		"GivenSingleEntry_WhenValueNumber_ThenReturnAsString": {
			givenKeysAndValues: map[string]interface{}{"key": 0},
			expectedResult:     "message \x1b[90m(key=\"0\")\x1b[0m",
		},
		"GivenSingleEntry_WhenValueArray_ThenReturnAsString": {
			givenKeysAndValues: map[string]interface{}{"key": []int{0, 1, 2}},
			expectedResult:     "message \x1b[90m(key=\"[0 1 2]\")\x1b[0m",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := DefaultFormatter("message", tt.givenKeysAndValues)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestPtermSink_WithName(t *testing.T) {
	sink := NewPtermSink()
	logger := logr.New(sink)

	t.Run("1st level scope", func(t *testing.T) {
		require.Equal(t, "", logger.GetSink().(PtermSink).scope)

		newLogger := logger.WithName("scope")
		newScope := newLogger.GetSink().(PtermSink).scope
		assert.NotEqual(t, logger.GetSink().(PtermSink).scope, newScope)
		assert.Equal(t, "scope", newScope)
	})
	t.Run("2nd level scope", func(t *testing.T) {
		require.Equal(t, "", logger.GetSink().(PtermSink).scope)

		newLogger := logger.WithName("scope").WithName("nested")
		newScope := newLogger.GetSink().(PtermSink).scope
		assert.NotEqual(t, logger.GetSink().(PtermSink).scope, newScope)
		assert.Equal(t, "scope:nested", newScope)
	})
}

func TestPtermSink_Enabled(t *testing.T) {
	sink := NewPtermSink()
	t.Run("GivenEnabledLevel_ThenReturnTrue", func(t *testing.T) {
		enabled := sink.Enabled(0)
		assert.True(t, enabled)
	})
	t.Run("GivenNonExistingLevel_ThenReturnFalse", func(t *testing.T) {
		enabled := sink.Enabled(10000)
		assert.False(t, enabled)
	})
	t.Run("GivenNonExistingLevel_WhenFallbackPrinterDefined_ThenReturnTrue", func(t *testing.T) {
		sink = *sink.WithFallbackPrinter(pterm.Debug)
		enabled := sink.Enabled(10000)
		assert.True(t, enabled)
	})
}

func TestPtermSink_WithValues(t *testing.T) {
	sink := NewPtermSink()

	sink.keyValues["key"] = "value"
	assert.Len(t, sink.keyValues, 1)

	t.Run("GivenNewInstance_ThenDoNotModifyExisting", func(t *testing.T) {
		sink.WithValues("foo", "bar")

		// Assert that it didn't modify existing instance
		assert.Equal(t, "value", sink.keyValues["key"])
		assert.Len(t, sink.keyValues, 1)
	})
	t.Run("GivenNewInstance_ThenCopyFromExistingInstance", func(t *testing.T) {
		newSink := sink.WithValues("foo", "bar").(*PtermSink)

		assert.Equal(t, "value", newSink.keyValues["key"])
		assert.Equal(t, "bar", newSink.keyValues["foo"])
		assert.Len(t, newSink.keyValues, 2)
	})
}

func TestPtermSink_toMap(t *testing.T) {
	tests := map[string]struct {
		givenKvs    []interface{}
		existingKvs map[string]interface{}
		expectedMap map[string]interface{}
	}{
		"GivenNilKeys_ThenReturnEmptyMap": {
			givenKvs:    nil,
			expectedMap: map[string]interface{}{},
		},
		"GivenEmptyKeys_ThenReturnEmptyMap": {
			givenKvs:    []interface{}{},
			expectedMap: map[string]interface{}{},
		},
		"GivenKeyWithValue_ThenReturnSinglyEntry": {
			givenKvs: []interface{}{"key", "value"},
			expectedMap: map[string]interface{}{
				"key": "value",
			},
		},
		"GivenKeyWithValue_WhenValueEmpty_ThenReturnSinglyEntry": {
			givenKvs: []interface{}{"key"},
			expectedMap: map[string]interface{}{
				"key": nil,
			},
		},
		"GivenExistingKeyWithValue_ThenReturnCombinedEntries": {
			givenKvs: []interface{}{"key"},
			existingKvs: map[string]interface{}{
				"foo": "bar",
			},
			expectedMap: map[string]interface{}{
				"key": nil,
				"foo": "bar",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sink := NewPtermSink()
			sink.keyValues = tt.existingKvs
			result := sink.toMap(tt.givenKvs...)

			assert.Equal(t, tt.expectedMap, result)
		})
	}
}

func TestPtermSink_WithOutput(t *testing.T) {
	old, out := bytes.Buffer{}, bytes.Buffer{}

	sink := NewPtermSink().WithOutput(&old)
	newSink := sink.WithOutput(&out)

	sink.Info(0, "shouldn't be included in new sink")
	newSink.Info(0, "message", "key", "value")

	// The expected output is actually from "golden" execution, but it should notify us on unnoticed changes
	assert.Equal(t, "\x1b[30;46m\x1b[30;46m  INFO   \x1b[0m\x1b[0m \x1b[96m\x1b[96mshouldn't be included in new sink\x1b[0m\x1b[0m\n", old.String(), "old sink")
	assert.Equal(t, "\x1b[30;46m\x1b[30;46m  INFO   \x1b[0m\x1b[0m \x1b[96m\x1b[96mmessage \x1b[90m(key=\"value\")\x1b[0m\x1b[96m\x1b[0m\x1b[0m\n", out.String(), "new sink")
}

func TestPtermSink_SetOutput(t *testing.T) {
	out := bytes.Buffer{}

	sink := NewPtermSink()
	newSink := sink.SetOutput(&out)

	sink.Info(0, "message", "key", "value")
	newSink.Info(0, "message", "key", "value")

	expected := "\x1b[30;46m\x1b[30;46m  INFO   \x1b[0m\x1b[0m \x1b[96m\x1b[96mmessage \x1b[90m(key=\"value\")\x1b[0m\x1b[96m\x1b[0m\x1b[0m\n"
	doubleLine := expected + expected
	// The expected output is actually from "golden" execution, but it should notify us on unnoticed changes
	assert.Equal(t, doubleLine, out.String())
}

func TestPtermSink_WithLevelEnabled(t *testing.T) {
	out := &bytes.Buffer{}

	rootSink := NewPtermSink()
	rootSink.SetOutput(out)

	rootLogger := logr.New(rootSink.SetLevelEnabled(1, false))
	rootLogger.Info("info message")
	rootLogger.V(1).Info("debug message")

	assert.Contains(t, out.String(), "info message")
	assert.NotContains(t, out.String(), "debug message")
}
