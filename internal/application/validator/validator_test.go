package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Required(t *testing.T) {
	// A struct with various types to test the 'Required' rule against.
	type testStruct struct {
		NonEmptyString string
		EmptyString    string
		NonZeroInt     int
		ZeroInt        int
		TrueBool       bool
		FalseBool      bool // The zero value for a bool is 'false'
		NonNilSlice    []string
		NilSlice       []string
		NonNilPtr      *int
		NilPtr         *int
	}

	// Initialize a pointer for the non-nil case
	nonNilVal := 5

	sample := testStruct{
		NonEmptyString: "hello",
		EmptyString:    "",
		NonZeroInt:     10,
		ZeroInt:        0,
		TrueBool:       true,
		FalseBool:      false,
		NonNilSlice:    make([]string, 1),
		NilSlice:       nil,
		NonNilPtr:      &nonNilVal,
		NilPtr:         nil,
	}

	testCases := []struct {
		name        string
		fieldName   string
		expectError bool
	}{
		{"should pass for non-empty string", "NonEmptyString", false},
		{"should fail for empty string", "EmptyString", true},
		{"should pass for non-zero int", "NonZeroInt", false},
		{"should fail for zero int", "ZeroInt", true},
		{"should pass for true bool", "TrueBool", false},
		{"should fail for false bool", "FalseBool", true},
		{"should pass for non-nil slice", "NonNilSlice", false},
		{"should fail for nil slice", "NilSlice", true},
		{"should pass for non-nil pointer", "NonNilPtr", false},
		{"should fail for nil pointer", "NilPtr", true},
		{"should fail for non-existent field", "InvalidField", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			v := New().For(sample)

			// Act
			v.Required(tc.fieldName)
			errs := v.Validate()

			// Assert
			if tc.expectError {
				assert.NotNil(t, errs)
				assert.Len(t, errs, 1, "should have exactly one error")
			} else {
				assert.Nil(t, errs)
			}
		})
	}
}

func TestValidator_NumberMin(t *testing.T) {
	type testStruct struct {
		IntField    int
		FloatField  float64
		StringField string
	}

	sample := testStruct{
		IntField:    10,
		FloatField:  10.5,
		StringField: "not a number",
	}

	testCases := []struct {
		name        string
		fieldName   string
		minValue    float64
		expectError bool
	}{
		{"should pass for int greater than min", "IntField", 5, false},
		{"should pass for int equal to min", "IntField", 10, false},
		{"should fail for int less than min", "IntField", 15, true},
		{"should pass for float greater than min", "FloatField", 5.5, false},
		{"should pass for float equal to min", "FloatField", 10.5, false},
		{"should fail for float less than min", "FloatField", 15.5, true},
		{"should fail for non-numeric field type", "StringField", 0, true},
		{"should fail for non-existent field", "InvalidField", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New().For(sample)
			v.NumberMin(tc.minValue, tc.fieldName)
			errs := v.Validate()

			if tc.expectError {
				assert.NotNil(t, errs)
			} else {
				assert.Nil(t, errs)
			}
		})
	}
}

func TestValidator_ChainingAndErrorHandling(t *testing.T) {
	type testStruct struct {
		Name  string
		Age   int
		Score float64
	}

	t.Run("should accumulate multiple errors correctly", func(t *testing.T) {
		// Arrange
		sample := testStruct{
			Name:  "",   // Fails Required
			Age:   17,   // Fails NumberMin(18)
			Score: -5.0, // Fails NumberMin(0)
		}

		// Act
		v := New().For(sample).
			Required("Name").
			NumberMin(18, "Age").
			NumberMin(0, "Score")

		// Assert for Validate()
		errs := v.Validate()
		require.NotNil(t, errs)
		assert.Len(t, errs, 3, "should have accumulated three errors")

		// Assert for ConsolidateErrors()
		consolidatedErr := v.ConsolidateErrors()
		require.Error(t, consolidatedErr)
		errorString := consolidatedErr.Error()
		assert.True(t, strings.Contains(errorString, "el campo Name es requerido"))
		assert.True(t, strings.Contains(errorString, "Age debe ser almenos 18"))
		assert.True(t, strings.Contains(errorString, "Score debe ser almenos 0"))
	})

	t.Run("should return nil when all validations pass", func(t *testing.T) {
		// Arrange
		sample := testStruct{
			Name:  "Valid Name",
			Age:   25,
			Score: 100,
		}

		// Act
		v := New().For(&sample). // Test with a pointer to a struct
						Required("Name", "Age", "Score").
						NumberMin(18, "Age").
						NumberMin(0, "Score")

		// Assert
		assert.Nil(t, v.Validate())
		assert.NoError(t, v.ConsolidateErrors())
	})

	t.Run("should handle invalid setup", func(t *testing.T) {
		v1 := New() // .For() is not called
		errs1 := v1.Required("any").Validate()
		require.NotNil(t, errs1)
		assert.Equal(t, "validation target not set. Call For() first", errs1[0].Error())

		v2 := New().For(123) // Target is not a struct
		errs2 := v2.Required("any").Validate()
		require.NotNil(t, errs2)
		assert.Equal(t, "validation target must be a struct", errs2[0].Error())
	})
}
