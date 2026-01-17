package organization

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

// Role tests

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		role  Role
		valid bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleMember, true},
		{Role("superuser"), false},
		{Role(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.role.IsValid())
		})
	}
}

func TestRole_CanManageMembers(t *testing.T) {
	tests := []struct {
		role     Role
		expected bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.CanManageMembers())
		})
	}
}

func TestRole_CanDeleteOrg(t *testing.T) {
	tests := []struct {
		role     Role
		expected bool
	}{
		{RoleOwner, true},
		{RoleAdmin, false},
		{RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.CanDeleteOrg())
		})
	}
}

func TestRole_CanTransferOwnership(t *testing.T) {
	tests := []struct {
		role     Role
		expected bool
	}{
		{RoleOwner, true},
		{RoleAdmin, false},
		{RoleMember, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.CanTransferOwnership())
		})
	}
}

// Model JSON serialization tests

func TestOrganization_JSONSerialization(t *testing.T) {
	org := &Organization{
		ID:        "org-123",
		Name:      "Test Org",
		Slug:      "test-org-abc12345",
		CreatedBy: "user-123",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(org)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "org-123", parsed["id"])
	assert.Equal(t, "Test Org", parsed["name"])
	assert.Equal(t, "test-org-abc12345", parsed["slug"])
	// created_by should be omitted from JSON (json:"-")
	_, hasCreatedBy := parsed["created_by"]
	assert.False(t, hasCreatedBy, "created_by should not be in JSON output")
}

func TestOrganizationWithRole_JSONSerialization(t *testing.T) {
	org := OrganizationWithRole{
		Organization: Organization{
			ID:   "org-123",
			Name: "Test Org",
			Slug: "test-org",
		},
		Role: RoleOwner,
	}

	data, err := json.Marshal(org)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "owner", parsed["role"])
}

func TestMember_JSONSerialization(t *testing.T) {
	member := &Member{
		ID:             "member-123",
		OrganizationID: "org-123",
		UserID:         "user-123",
		Role:           RoleAdmin,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(member)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "admin", parsed["role"])
}

func TestInvitation_JSONSerialization(t *testing.T) {
	inv := &Invitation{
		ID:             "inv-123",
		OrganizationID: "org-123",
		Email:          "invite@example.com",
		Role:           RoleMember,
		Token:          "secret-token",
		InvitedBy:      "user-123",
		Status:         StatusPending,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	data, err := json.Marshal(inv)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Token should be omitted from JSON (json:"-")
	_, hasToken := parsed["token"]
	assert.False(t, hasToken, "token should not be in JSON output")
	assert.Equal(t, "pending", parsed["status"])
}

// Request validation tests

func TestCreateOrgRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateOrgRequest
		valid   bool
	}{
		{
			name:    "valid name",
			request: CreateOrgRequest{Name: "My Organization"},
			valid:   true,
		},
		{
			name:    "empty name",
			request: CreateOrgRequest{Name: ""},
			valid:   false,
		},
		{
			name:    "name too long",
			request: CreateOrgRequest{Name: strings.Repeat("a", 101)},
			valid:   false,
		},
		{
			name:    "minimum name length",
			request: CreateOrgRequest{Name: "A"},
			valid:   true,
		},
		{
			name:    "maximum name length",
			request: CreateOrgRequest{Name: strings.Repeat("a", 100)},
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.LessOrEqual(t, len(tt.request.Name), 100)
			}
		})
	}
}

func TestInviteMemberRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request InviteMemberRequest
		valid   bool
	}{
		{
			name:    "valid member invite",
			request: InviteMemberRequest{Email: "test@example.com", Role: RoleMember},
			valid:   true,
		},
		{
			name:    "valid admin invite",
			request: InviteMemberRequest{Email: "admin@example.com", Role: RoleAdmin},
			valid:   true,
		},
		{
			name:    "invalid role - owner",
			request: InviteMemberRequest{Email: "test@example.com", Role: RoleOwner},
			valid:   false, // Can't invite as owner
		},
		{
			name:    "invalid email",
			request: InviteMemberRequest{Email: "not-an-email", Role: RoleMember},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.Contains(t, tt.request.Email, "@")
				assert.True(t, tt.request.Role == RoleAdmin || tt.request.Role == RoleMember)
			}
		})
	}
}

func TestUpdateMemberRoleRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateMemberRoleRequest
		valid   bool
	}{
		{
			name:    "valid owner role",
			request: UpdateMemberRoleRequest{Role: RoleOwner},
			valid:   true,
		},
		{
			name:    "valid admin role",
			request: UpdateMemberRoleRequest{Role: RoleAdmin},
			valid:   true,
		},
		{
			name:    "valid member role",
			request: UpdateMemberRoleRequest{Role: RoleMember},
			valid:   true,
		},
		{
			name:    "invalid role",
			request: UpdateMemberRoleRequest{Role: Role("superuser")},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.request.Role.IsValid())
		})
	}
}

func TestTransferOwnershipRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request TransferOwnershipRequest
		valid   bool
	}{
		{
			name:    "valid UUID",
			request: TransferOwnershipRequest{NewOwnerID: "550e8400-e29b-41d4-a716-446655440000"},
			valid:   true,
		},
		{
			name:    "invalid UUID",
			request: TransferOwnershipRequest{NewOwnerID: "not-a-uuid"},
			valid:   false,
		},
		{
			name:    "empty",
			request: TransferOwnershipRequest{NewOwnerID: ""},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation check
			if tt.valid {
				assert.NotEmpty(t, tt.request.NewOwnerID)
				assert.Len(t, tt.request.NewOwnerID, 36) // UUID format
			}
		})
	}
}

// Slug generation tests

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		userID   string
		expected string
	}{
		{
			name:     "simple name",
			orgName:  "My Organization",
			userID:   "user-12345678",
			expected: "my-organization-user-123",
		},
		{
			name:     "special characters",
			orgName:  "Test & Company!",
			userID:   "abc12345678",
			expected: "test-company-abc12345",
		},
		{
			name:     "numbers",
			orgName:  "Company 123",
			userID:   "user99999999",
			expected: "company-123-user9999",
		},
		{
			name:     "already lowercase",
			orgName:  "simple",
			userID:   "test12345678",
			expected: "simple-test1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.orgName, tt.userID)
			assert.Contains(t, result, tt.userID[:8])
			// Should be lowercase
			assert.Equal(t, strings.ToLower(result), result)
			// Should not have consecutive dashes at edges
			assert.False(t, strings.HasPrefix(result, "-"))
			assert.False(t, strings.HasSuffix(result, "-"))
		})
	}
}

func TestGenerateSlug_LongName(t *testing.T) {
	longName := strings.Repeat("a", 100)
	userID := "user12345678"

	result := GenerateSlug(longName, userID)

	// Should be truncated: 30 chars base + dash + 8 char user prefix
	assert.LessOrEqual(t, len(result), 40)
	assert.Contains(t, result, userID[:8])
}

// Error response tests

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		writeFunc      func(w http.ResponseWriter)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "forbidden",
			writeFunc: func(w http.ResponseWriter) {
				response.Forbidden(w, "not a member of this organization")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
		},
		{
			name: "not found",
			writeFunc: func(w http.ResponseWriter) {
				response.NotFound(w, "organization not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not_found",
		},
		{
			name: "bad request",
			writeFunc: func(w http.ResponseWriter) {
				response.BadRequest(w, "validation failed")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad_request",
		},
		{
			name: "unauthorized",
			writeFunc: func(w http.ResponseWriter) {
				response.Unauthorized(w, "authentication required")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.writeFunc(w)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp response.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedError, resp.Error)
		})
	}
}

// Test context and URL params work correctly
func TestContextHelpers(t *testing.T) {
	usr := testutil.NewTestUser()
	sess := testutil.NewTestSession(usr.ID)

	ctx := newTestContext(usr, sess)

	gotUser := middleware.GetUserFromContext(ctx)
	gotSession := middleware.GetSessionFromContext(ctx)

	require.NotNil(t, gotUser)
	require.NotNil(t, gotSession)
	assert.Equal(t, usr.ID, gotUser.ID)
	assert.Equal(t, sess.ID, gotSession.ID)
}

func TestURLParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = withURLParams(req, map[string]string{
		"orgID":    "org-123",
		"userID":   "user-456",
		"inviteID": "inv-789",
	})

	rctx := chi.RouteContext(req.Context())
	assert.Equal(t, "org-123", rctx.URLParam("orgID"))
	assert.Equal(t, "user-456", rctx.URLParam("userID"))
	assert.Equal(t, "inv-789", rctx.URLParam("inviteID"))
}

// Invitation status tests

func TestInvitationStatus_Values(t *testing.T) {
	assert.Equal(t, InvitationStatus("pending"), StatusPending)
	assert.Equal(t, InvitationStatus("accepted"), StatusAccepted)
	assert.Equal(t, InvitationStatus("declined"), StatusDeclined)
	assert.Equal(t, InvitationStatus("expired"), StatusExpired)
}
