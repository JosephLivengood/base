package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           any
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "with data",
			status:         http.StatusOK,
			data:           map[string]string{"key": "value"},
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
		{
			name:           "with nil data",
			status:         http.StatusOK,
			data:           nil,
			expectedStatus: http.StatusOK,
			checkBody:      false,
		},
		{
			name:           "custom status code",
			status:         http.StatusCreated,
			data:           map[string]string{"id": "123"},
			expectedStatus: http.StatusCreated,
			checkBody:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			JSON(w, tt.status, tt.data)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			if tt.checkBody {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}
	OK(w, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123", "name": "test"}
	Created(w, data)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad_request", "invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "bad_request", resp.Error)
	assert.Equal(t, "invalid input", resp.Message)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	BadRequest(w, "validation failed")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "bad_request", resp.Error)
	assert.Equal(t, "validation failed", resp.Message)
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "authentication required")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
	assert.Equal(t, "authentication required", resp.Message)
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w, "access denied")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
	assert.Equal(t, "access denied", resp.Message)
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "not_found", resp.Error)
	assert.Equal(t, "resource not found", resp.Message)
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "something went wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "internal_error", resp.Error)
	assert.Equal(t, "something went wrong", resp.Message)
}
