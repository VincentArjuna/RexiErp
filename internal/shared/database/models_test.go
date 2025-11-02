package database

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBaseModel_BeforeCreate(t *testing.T) {
	t.Run("NewModel", func(t *testing.T) {
		model := &BaseModel{}
		err := model.BeforeCreate(nil)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, model.ID)
		assert.False(t, model.CreatedAt.IsZero())
		assert.False(t, model.UpdatedAt.IsZero())
	})

	t.Run("ExistingModel", func(t *testing.T) {
		existingTime := time.Now().Add(-1 * time.Hour)
		model := &BaseModel{
			ID:        uuid.New(),
			CreatedAt: existingTime,
			UpdatedAt: existingTime,
		}

		err := model.BeforeCreate(nil)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, model.ID)
		assert.Equal(t, existingTime, model.CreatedAt) // Should not change
		assert.False(t, model.UpdatedAt.IsZero())      // Should be updated
	})
}

func TestBaseModel_BeforeUpdate(t *testing.T) {
	model := &BaseModel{
		ID:        uuid.New(),
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	err := model.BeforeUpdate(nil)

	assert.NoError(t, err)
	assert.True(t, model.UpdatedAt.After(model.CreatedAt))
}

func TestJSONB_Value(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var j JSONB
		value, err := j.Value()

		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("ValidValue", func(t *testing.T) {
		j := JSONB{"key": "value", "number": 42}
		value, err := j.Value()

		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Contains(t, string(value.([]byte)), "key")
		assert.Contains(t, string(value.([]byte)), "value")
	})
}

func TestJSONB_Scan(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var j JSONB
		err := j.Scan(nil)

		assert.NoError(t, err)
		assert.Nil(t, j)
	})

	t.Run("ByteSlice", func(t *testing.T) {
		var j JSONB
		data := []byte(`{"key":"value"}`)
		err := j.Scan(data)

		assert.NoError(t, err)
		assert.Equal(t, "value", j["key"])
	})

	t.Run("String", func(t *testing.T) {
		var j JSONB
		data := `{"key":"value"}`
		err := j.Scan(data)

		assert.NoError(t, err)
		assert.Equal(t, "value", j["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		var j JSONB
		data := `invalid json`
		err := j.Scan(data)

		assert.Error(t, err)
	})
}

func TestStringArray_Value(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var s StringArray
		value, err := s.Value()

		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("ValidValue", func(t *testing.T) {
		s := StringArray{"item1", "item2", "item3"}
		value, err := s.Value()

		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, s, value)
	})
}

func TestStringArray_Scan(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var s StringArray
		err := s.Scan(nil)

		assert.NoError(t, err)
		assert.Nil(t, s)
	})

	t.Run("ValidArray", func(t *testing.T) {
		var s StringArray
		data := []byte(`["item1","item2","item3"]`)
		err := s.Scan(data)

		assert.NoError(t, err)
		assert.Equal(t, StringArray{"item1", "item2", "item3"}, s)
	})
}

func TestValidator_Required(t *testing.T) {
	t.Run("ValidValue", func(t *testing.T) {
		v := NewValidator()
		v.Required("field", "value")

		assert.False(t, v.HasErrors())
	})

	t.Run("EmptyString", func(t *testing.T) {
		v := NewValidator()
		v.Required("field", "")

		assert.True(t, v.HasErrors())
		assert.Equal(t, "field", v.Errors()[0].Field)
	})

	t.Run("NilValue", func(t *testing.T) {
		v := NewValidator()
		v.Required("field", nil)

		assert.True(t, v.HasErrors())
		assert.Equal(t, "field", v.Errors()[0].Field)
	})
}

func TestValidator_MinLength(t *testing.T) {
	t.Run("ValidLength", func(t *testing.T) {
		v := NewValidator()
		v.MinLength("field", "abcdef", 5)

		assert.False(t, v.HasErrors())
	})

	t.Run("InvalidLength", func(t *testing.T) {
		v := NewValidator()
		v.MinLength("field", "abc", 5)

		assert.True(t, v.HasErrors())
	})
}

func TestValidator_MaxLength(t *testing.T) {
	t.Run("ValidLength", func(t *testing.T) {
		v := NewValidator()
		v.MaxLength("field", "abc", 5)

		assert.False(t, v.HasErrors())
	})

	t.Run("InvalidLength", func(t *testing.T) {
		v := NewValidator()
		v.MaxLength("field", "abcdef", 5)

		assert.True(t, v.HasErrors())
	})
}

func TestValidator_Email(t *testing.T) {
	testCases := []struct {
		email    string
		valid    bool
	}{
		{"", true},                  // Empty email should not trigger validation
		{"valid@example.com", true},
		{"invalid", false},
		{"invalid@", false},
		{"@example.com", false},
		{strings.Repeat("a", 300) + "@example.com", false}, // Too long
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			v := NewValidator()
			v.Email("email", tc.email)

			if tc.valid {
				assert.False(t, v.HasErrors())
			} else {
				assert.True(t, v.HasErrors())
			}
		})
	}
}

func TestValidator_UUID(t *testing.T) {
	testCases := []struct {
		id    string
		valid bool
	}{
		{"", true}, // Empty UUID should not trigger validation
		{uuid.New().String(), true},
		{"invalid-uuid", false},
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400-e29b-41d4-a716-44665544zzzz", false},
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			v := NewValidator()
			v.UUID("id", tc.id)

			if tc.valid {
				assert.False(t, v.HasErrors())
			} else {
				assert.True(t, v.HasErrors())
			}
		})
	}
}

func TestValidator_OneOf(t *testing.T) {
	t.Run("ValidValue", func(t *testing.T) {
		v := NewValidator()
		v.OneOf("status", "active", []string{"active", "inactive", "pending"})

		assert.False(t, v.HasErrors())
	})

	t.Run("InvalidValue", func(t *testing.T) {
		v := NewValidator()
		v.OneOf("status", "invalid", []string{"active", "inactive", "pending"})

		assert.True(t, v.HasErrors())
	})

	t.Run("EmptyValue", func(t *testing.T) {
		v := NewValidator()
		v.OneOf("status", "", []string{"active", "inactive", "pending"})

		assert.False(t, v.HasErrors()) // Empty should not trigger validation
	})
}

func TestValidator_Range(t *testing.T) {
	t.Run("InRange", func(t *testing.T) {
		v := NewValidator()
		v.Range("age", 25, 18, 65)

		assert.False(t, v.HasErrors())
	})

	t.Run("TooLow", func(t *testing.T) {
		v := NewValidator()
		v.Range("age", 15, 18, 65)

		assert.True(t, v.HasErrors())
	})

	t.Run("TooHigh", func(t *testing.T) {
		v := NewValidator()
		v.Range("age", 70, 18, 65)

		assert.True(t, v.HasErrors())
	})
}

func TestValidator_Positive(t *testing.T) {
	t.Run("PositiveNumber", func(t *testing.T) {
		v := NewValidator()
		v.Positive("count", 5)

		assert.False(t, v.HasErrors())
	})

	t.Run("Zero", func(t *testing.T) {
		v := NewValidator()
		v.Positive("count", 0)

		assert.True(t, v.HasErrors())
	})

	t.Run("NegativeNumber", func(t *testing.T) {
		v := NewValidator()
		v.Positive("count", -5)

		assert.True(t, v.HasErrors())
	})
}

func TestValidator_NonNegative(t *testing.T) {
	t.Run("PositiveNumber", func(t *testing.T) {
		v := NewValidator()
		v.NonNegative("count", 5)

		assert.False(t, v.HasErrors())
	})

	t.Run("Zero", func(t *testing.T) {
		v := NewValidator()
		v.NonNegative("count", 0)

		assert.False(t, v.HasErrors())
	})

	t.Run("NegativeNumber", func(t *testing.T) {
		v := NewValidator()
		v.NonNegative("count", -5)

		assert.True(t, v.HasErrors())
	})
}

func TestValidator_FutureDate(t *testing.T) {
	t.Run("FutureDate", func(t *testing.T) {
		v := NewValidator()
		future := time.Now().Add(1 * time.Hour)
		v.FutureDate("date", future)

		assert.False(t, v.HasErrors())
	})

	t.Run("PastDate", func(t *testing.T) {
		v := NewValidator()
		past := time.Now().Add(-1 * time.Hour)
		v.FutureDate("date", past)

		assert.True(t, v.HasErrors())
	})

	t.Run("ZeroDate", func(t *testing.T) {
		v := NewValidator()
		v.FutureDate("date", time.Time{})

		assert.False(t, v.HasErrors()) // Zero time should not trigger validation
	})
}

func TestValidator_PastDate(t *testing.T) {
	t.Run("PastDate", func(t *testing.T) {
		v := NewValidator()
		past := time.Now().Add(-1 * time.Hour)
		v.PastDate("date", past)

		assert.False(t, v.HasErrors())
	})

	t.Run("FutureDate", func(t *testing.T) {
		v := NewValidator()
		future := time.Now().Add(1 * time.Hour)
		v.PastDate("date", future)

		assert.True(t, v.HasErrors())
	})

	t.Run("ZeroDate", func(t *testing.T) {
		v := NewValidator()
		v.PastDate("date", time.Time{})

		assert.False(t, v.HasErrors()) // Zero time should not trigger validation
	})
}

func TestValidator_ToError(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		v := NewValidator()
		err := v.ToError()

		assert.NoError(t, err)
	})

	t.Run("SingleError", func(t *testing.T) {
		v := NewValidator()
		v.Required("field", "")
		err := v.ToError()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field")
	})

	t.Run("MultipleErrors", func(t *testing.T) {
		v := NewValidator()
		v.Required("field1", "")
		v.Required("field2", "")
		err := v.ToError()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multiple validation errors")
	})
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		var ves ValidationErrors
		err := ves.Error()

		assert.Equal(t, "no validation errors", err)
	})

	t.Run("SingleError", func(t *testing.T) {
		ves := ValidationErrors{
			{Field: "field1", Message: "is required"},
		}
		err := ves.Error()

		assert.Contains(t, err, "field1")
		assert.Contains(t, err, "is required")
	})

	t.Run("MultipleErrors", func(t *testing.T) {
		ves := ValidationErrors{
			{Field: "field1", Message: "is required"},
			{Field: "field2", Message: "is invalid"},
		}
		err := ves.Error()

		assert.Contains(t, err, "multiple validation errors")
	})
}

func TestValidationErrors_Add(t *testing.T) {
	var ves ValidationErrors
	ves.Add("field", "is required", "")

	assert.Equal(t, 1, len(ves))
	assert.Equal(t, "field", ves[0].Field)
	assert.Equal(t, "is required", ves[0].Message)
}

func TestValidationErrors_HasErrors(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		var ves ValidationErrors
		assert.False(t, ves.HasErrors())
	})

	t.Run("HasErrors", func(t *testing.T) {
		ves := ValidationErrors{{Field: "field", Message: "error"}}
		assert.True(t, ves.HasErrors())
	})
}

// Test utility functions
func TestGenerateUUID(t *testing.T) {
	id := GenerateUUID()
	assert.NotEqual(t, uuid.Nil, id)

	// Test that generated UUIDs are unique
	id2 := GenerateUUID()
	assert.NotEqual(t, id, id2)
}

func TestParseUUID(t *testing.T) {
	t.Run("ValidUUID", func(t *testing.T) {
		expected := uuid.New()
		parsed, err := ParseUUID(expected.String())

		assert.NoError(t, err)
		assert.Equal(t, expected, parsed)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		_, err := ParseUUID("invalid-uuid")
		assert.Error(t, err)
	})
}

func TestIsValidUUID(t *testing.T) {
	t.Run("ValidUUID", func(t *testing.T) {
		valid := uuid.New().String()
		assert.True(t, IsValidUUID(valid))
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		assert.False(t, IsValidUUID("invalid-uuid"))
		assert.False(t, IsValidUUID(""))
	})
}

func TestNowUTC(t *testing.T) {
	now := NowUTC()
	assert.False(t, now.IsZero())
	assert.Equal(t, time.UTC, now.Location())
}

func TestFormatTimestamp(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.Local)
	formatted := FormatTimestamp(testTime)

	assert.Equal(t, 2023, formatted.Year())
	assert.Equal(t, time.UTC, formatted.Location())
}

func TestParseTimestamp(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	parsed := ParseTimestamp(testTime)

	assert.Equal(t, testTime, parsed)
	assert.Equal(t, time.UTC, parsed.Location())
}

func TestCalculateHash(t *testing.T) {
	content := "test content"
	hash := CalculateHash(content)

	assert.NotEmpty(t, hash)
	assert.Equal(t, hash, CalculateHash(content)) // Should be consistent
}

func TestSanitizeInput(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"  valid input  ", "valid input"},
		{"", ""},
		{"input\x00with\x00nulls", "inputwithnulls"},
		{"正常输入", "正常输入"}, // Unicode should be preserved
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := SanitizeInput(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
	}{
		{`{"key": "value"}`, true},
		{`{"array": [1, 2, 3]}`, true},
		{`invalid json`, false},
		{``, false},
		{`null`, true},
		{`"string"`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := IsValidJSON(tc.input)
			assert.Equal(t, tc.valid, result)
		})
	}
}

func TestMaskSensitiveData(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"a", "****"},
		{"ab", "****"},
		{"abc", "a****c"},
		{"password123", "pa****23"},
		{"1234567890", "12****90"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := MaskSensitiveData(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

