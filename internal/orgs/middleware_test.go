package orgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/auth"
)

func TestOrgRoleMiddleware_NoMemberInContext(t *testing.T) {
	mw := OrgRoleMiddleware(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestOrgRoleMiddleware_InsufficientRole(t *testing.T) {
	mw := OrgRoleMiddleware(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	member := &Member{
		ID:     uuid.New(),
		OrgID:  uuid.New(),
		UserID: uuid.New(),
		Role:   RoleViewer,
	}
	ctx := SetMemberInContext(context.Background(), member)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestOrgRoleMiddleware_SufficientRole(t *testing.T) {
	mw := OrgRoleMiddleware(RoleAdmin)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	member := &Member{
		ID:     uuid.New(),
		OrgID:  uuid.New(),
		UserID: uuid.New(),
		Role:   RoleOwner,
	}
	ctx := SetMemberInContext(context.Background(), member)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestContextHelpers(t *testing.T) {
	org := &Organization{ID: uuid.New(), Name: "test"}
	member := &Member{ID: uuid.New(), Role: RoleAdmin}

	ctx := context.Background()

	// Nil when not set
	if GetOrgFromContext(ctx) != nil {
		t.Error("expected nil org from empty context")
	}
	if GetMemberFromContext(ctx) != nil {
		t.Error("expected nil member from empty context")
	}

	ctx = SetOrgInContext(ctx, org)
	ctx = SetMemberInContext(ctx, member)

	if got := GetOrgFromContext(ctx); got != org {
		t.Error("expected org from context")
	}
	if got := GetMemberFromContext(ctx); got != member {
		t.Error("expected member from context")
	}
}

// Ensure auth import is used — verify claims key type exists
func TestAuthClaimsTypeCheck(t *testing.T) {
	claims := &auth.AccessClaims{UserID: "test"}
	if claims.UserID != "test" {
		t.Error("unexpected")
	}
}
