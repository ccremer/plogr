package plogr

import (
	"testing"

	"github.com/go-logr/logr"
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
		newScope := newLogger.GetSink().(*PtermSink).scope
		assert.NotEqual(t, logger.GetSink().(PtermSink).scope, newScope)
		assert.Equal(t, "scope", newScope)
	})
	t.Run("2nd level scope", func(t *testing.T) {
		require.Equal(t, "", logger.GetSink().(PtermSink).scope)

		newLogger := logger.WithName("scope").WithName("nested")
		newScope := newLogger.GetSink().(*PtermSink).scope
		assert.NotEqual(t, logger.GetSink().(PtermSink).scope, newScope)
		assert.Equal(t, "scope:nested", newScope)
	})
}
