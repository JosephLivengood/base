package validate

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type testUser struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

type testRole struct {
	Role string `json:"role" validate:"required,oneof=admin member guest"`
}

type testURL struct {
	Website string `json:"website" validate:"required,url"`
}

type testUUID struct {
	ItemId string `json:"item_id" validate:"required,uuid"`
}

func TestStruct_ValidInput(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "valid user",
			input: testUser{Name: "John Doe", Email: "john@example.com", Age: 30},
		},
		{
			name:  "valid role admin",
			input: testRole{Role: "admin"},
		},
		{
			name:  "valid role member",
			input: testRole{Role: "member"},
		},
		{
			name:  "valid url",
			input: testURL{Website: "https://example.com"},
		},
		{
			name:  "valid uuid",
			input: testUUID{ItemId: "550e8400-e29b-41d4-a716-446655440000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestStruct_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		expectedField string
		expectedMsg   string
	}{
		{
			name:          "missing required name",
			input:         testUser{Email: "john@example.com"},
			expectedField: "name",
			expectedMsg:   "is required",
		},
		{
			name:          "invalid email",
			input:         testUser{Name: "John", Email: "invalid-email"},
			expectedField: "email",
			expectedMsg:   "must be a valid email address",
		},
		{
			name:          "name too short",
			input:         testUser{Name: "J", Email: "john@example.com"},
			expectedField: "name",
			expectedMsg:   "must be at least 2 characters",
		},
		{
			name:          "name too long",
			input:         testUser{Name: strings.Repeat("a", 51), Email: "john@example.com"},
			expectedField: "name",
			expectedMsg:   "must be at most 50 characters",
		},
		{
			name:          "invalid role",
			input:         testRole{Role: "superuser"},
			expectedField: "role",
			expectedMsg:   "must be one of: admin member guest",
		},
		{
			name:          "invalid url",
			input:         testURL{Website: "not-a-url"},
			expectedField: "website",
			expectedMsg:   "must be a valid URL",
		},
		{
			name:          "invalid uuid",
			input:         testUUID{ItemId: "not-a-uuid"},
			expectedField: "item_id",
			expectedMsg:   "must be a valid UUID",
		},
		{
			name:          "age below minimum",
			input:         testUser{Name: "John", Email: "john@example.com", Age: -1},
			expectedField: "age",
			expectedMsg:   "must be at least 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(tt.input)
			require.Error(t, err)

			var ve ValidationErrors
			require.ErrorAs(t, err, &ve)
			require.NotEmpty(t, ve.Errors)

			found := false
			for _, e := range ve.Errors {
				if e.Field == tt.expectedField && strings.Contains(e.Message, tt.expectedMsg) {
					found = true
					break
				}
			}
			assert.True(t, found, "expected field %s with message containing %q, got %v", tt.expectedField, tt.expectedMsg, ve.Errors)
		})
	}
}

func TestRequest_ValidJSON(t *testing.T) {
	body := `{"name": "John Doe", "email": "john@example.com", "age": 30}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var user testUser
	err := Request(req, &user)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, 30, user.Age)
}

func TestRequest_InvalidJSON(t *testing.T) {
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var user testUser
	err := Request(req, &user)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestRequest_ValidationFailure(t *testing.T) {
	body := `{"name": "", "email": "invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var user testUser
	err := Request(req, &user)

	require.Error(t, err)
	var ve ValidationErrors
	require.ErrorAs(t, err, &ve)
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"HTMLParser", "h_t_m_l_parser"},
		{"userID", "user_i_d"},
		{"already_snake", "already_snake"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFirstError(t *testing.T) {
	t.Run("with validation errors", func(t *testing.T) {
		user := testUser{Name: "", Email: "invalid"}
		err := Struct(user)
		require.Error(t, err)

		msg := FirstError(err)
		assert.NotEmpty(t, msg)
		// Should contain field name and message
		assert.Contains(t, msg, " ")
	})

	t.Run("with non-validation error", func(t *testing.T) {
		err := assert.AnError
		msg := FirstError(err)
		assert.Equal(t, err.Error(), msg)
	})
}

func TestValidationErrors_Error(t *testing.T) {
	ve := ValidationErrors{
		Errors: []ValidationError{
			{Field: "name", Message: "is required"},
			{Field: "email", Message: "must be a valid email address"},
		},
	}

	errMsg := ve.Error()
	assert.Contains(t, errMsg, "name: is required")
	assert.Contains(t, errMsg, "email: must be a valid email address")
	assert.Contains(t, errMsg, ";")
}
