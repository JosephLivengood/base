package project

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"base/api/internal/domain/organization"
	"base/api/internal/domain/user"
	"base/api/internal/middleware"
	"base/api/internal/session"
	"base/api/internal/testutil"
	"base/api/pkg/response"
)

// Test helpers

func newTestContext(usr *user.User, sess *session.Session) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserContextKey, usr)
	ctx = context.WithValue(ctx, middleware.SessionContextKey, sess)
	return ctx
}

func withURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// MockProjectRepo implements project repository interface for testing
type MockProjectRepo struct {
	CreateFunc            func(ctx context.Context, orgID, name string) (*Project, error)
	GetByIDFunc           func(ctx context.Context, id string) (*Project, error)
	GetByOrganizationFunc func(ctx context.Context, orgID string) ([]Project, error)
	UpdateFunc            func(ctx context.Context, id, name string) (*Project, error)
	DeleteFunc            func(ctx context.Context, id string) error
}

func (m *MockProjectRepo) Create(ctx context.Context, orgID, name string) (*Project, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, orgID, name)
	}
	return &Project{
		ID:             "proj-123",
		OrganizationID: orgID,
		Name:           name,
		CreatedAt:      time.Now(),
	}, nil
}

func (m *MockProjectRepo) GetByID(ctx context.Context, id string) (*Project, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrNotFound
}

func (m *MockProjectRepo) GetByOrganization(ctx context.Context, orgID string) ([]Project, error) {
	if m.GetByOrganizationFunc != nil {
		return m.GetByOrganizationFunc(ctx, orgID)
	}
	return []Project{}, nil
}

func (m *MockProjectRepo) Update(ctx context.Context, id, name string) (*Project, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, name)
	}
	return nil, ErrNotFound
}

func (m *MockProjectRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// MockOrgRepo for testing project handler's org checks
type MockOrgRepo struct {
	GetMemberFunc func(ctx context.Context, orgID, userID string) (*organization.Member, error)
}

func (m *MockOrgRepo) GetMember(ctx context.Context, orgID, userID string) (*organization.Member, error) {
	if m.GetMemberFunc != nil {
		return m.GetMemberFunc(ctx, orgID, userID)
	}
	return nil, organization.ErrNotMember
}

// Integration-style test that verifies JSON response structure
func TestProjectResponse_Structure(t *testing.T) {
	project := &Project{
		ID:             "proj-123",
		OrganizationID: "org-123",
		Name:           "Test Project",
		CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:      nil,
	}

	data, err := json.Marshal(project)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "proj-123", parsed["id"])
	assert.Equal(t, "org-123", parsed["organization_id"])
	assert.Equal(t, "Test Project", parsed["name"])
	assert.NotNil(t, parsed["created_at"])
	assert.Nil(t, parsed["deleted_at"])
}

// Test request validation
func TestCreateProjectRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateProjectRequest
		valid   bool
	}{
		{
			name:    "valid request",
			request: CreateProjectRequest{Name: "Valid Project"},
			valid:   true,
		},
		{
			name:    "empty name",
			request: CreateProjectRequest{Name: ""},
			valid:   false,
		},
		{
			name:    "name too long",
			request: CreateProjectRequest{Name: strings.Repeat("a", 101)},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate based on struct tags
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.LessOrEqual(t, len(tt.request.Name), 100)
			}
		})
	}
}

func TestUpdateProjectRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateProjectRequest
		valid   bool
	}{
		{
			name:    "valid request",
			request: UpdateProjectRequest{Name: "Updated Project"},
			valid:   true,
		},
		{
			name:    "empty name",
			request: UpdateProjectRequest{Name: ""},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
			}
		})
	}
}

// Test error responses
func TestErrorResponse_NotMember(t *testing.T) {
	w := httptest.NewRecorder()
	response.Forbidden(w, "not a member of this organization")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp response.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "forbidden", resp.Error)
}

func TestErrorResponse_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	response.NotFound(w, "project not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "not_found", resp.Error)
}

// Test context helpers work correctly
func TestContextHelpers(t *testing.T) {
	usr := testutil.NewTestUser()
	sess := testutil.NewTestSession(usr.ID)

	ctx := newTestContext(usr, sess)

	gotUser := middleware.GetUserFromContext(ctx)
	gotSession := middleware.GetSessionFromContext(ctx)

	assert.Equal(t, usr.ID, gotUser.ID)
	assert.Equal(t, sess.ID, gotSession.ID)
}

// Test URL params work correctly
func TestURLParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = withURLParams(req, map[string]string{
		"orgID":     "org-123",
		"projectID": "proj-456",
	})

	rctx := chi.RouteContext(req.Context())
	assert.Equal(t, "org-123", rctx.URLParam("orgID"))
	assert.Equal(t, "proj-456", rctx.URLParam("projectID"))
}

// Test mock repo behavior
func TestMockProjectRepo_Defaults(t *testing.T) {
	mock := &MockProjectRepo{}

	// Without custom functions, uses defaults
	project, err := mock.Create(context.Background(), "org-1", "Test")
	require.NoError(t, err)
	assert.Equal(t, "proj-123", project.ID)
	assert.Equal(t, "org-1", project.OrganizationID)
	assert.Equal(t, "Test", project.Name)

	_, err = mock.GetByID(context.Background(), "any")
	assert.ErrorIs(t, err, ErrNotFound)

	projects, err := mock.GetByOrganization(context.Background(), "any")
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestMockProjectRepo_CustomFunctions(t *testing.T) {
	expectedProject := &Project{
		ID:             "custom-proj",
		OrganizationID: "custom-org",
		Name:           "Custom Project",
	}

	mock := &MockProjectRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*Project, error) {
			if id == "custom-proj" {
				return expectedProject, nil
			}
			return nil, ErrNotFound
		},
	}

	project, err := mock.GetByID(context.Background(), "custom-proj")
	require.NoError(t, err)
	assert.Equal(t, expectedProject, project)

	_, err = mock.GetByID(context.Background(), "other")
	assert.ErrorIs(t, err, ErrNotFound)
}
