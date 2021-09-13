package plogr

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPtermSink_KvPairs(t *testing.T) {
	tests := map[string]struct {
		givenKVMap         map[string]interface{}
		givenKeysAndValues []interface{}
		expectedResult     []string
	}{
		"GivenNil_ThenReturnEmptyString": {
			givenKeysAndValues: nil,
			expectedResult:     []string{},
		},
		"GivenEmptyList_ThenReturnEmptyString": {
			givenKeysAndValues: []interface{}{},
			expectedResult:     []string{},
		},
		"GivenSingleEntry_WhenValueNil_ThenReturnNilRepresentation": {
			givenKeysAndValues: []interface{}{"key"},
			expectedResult:     []string{"key=\"<nil>\""},
		},
		"GivenSingleEntry_WhenValueNumber_ThenReturnAsString": {
			givenKeysAndValues: []interface{}{"key", 0},
			expectedResult:     []string{"key=\"0\""},
		},
		"GivenSingleEntry_WhenValueArray_ThenReturnAsString": {
			givenKeysAndValues: []interface{}{"key", []int{0, 1, 2}},
			expectedResult:     []string{"key=\"[0 1 2]\""},
		},
		"GivenExistingKeyValueMap_WhenKVGiven_ThenCombine": {
			givenKVMap:         map[string]interface{}{"another": "value"},
			givenKeysAndValues: []interface{}{"key", []int{0, 1, 2}},
			expectedResult:     []string{"another=\"value\"", "key=\"[0 1 2]\""},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewPtermSink()
			s.keyValues = tt.givenKVMap
			result := s.kvPairs(tt.givenKeysAndValues...)
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
