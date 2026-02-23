package orgs

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
)

type contextKey string

const (
	orgCtxKey    contextKey = "org"
	memberCtxKey contextKey = "org_member"
)

func SetOrgInContext(ctx context.Context, org *Organization) context.Context {
	return context.WithValue(ctx, orgCtxKey, org)
}

func GetOrgFromContext(ctx context.Context) *Organization {
	org, _ := ctx.Value(orgCtxKey).(*Organization)
	return org
}

func SetMemberInContext(ctx context.Context, member *Member) context.Context {
	return context.WithValue(ctx, memberCtxKey, member)
}

func GetMemberFromContext(ctx context.Context) *Member {
	member, _ := ctx.Value(memberCtxKey).(*Member)
	return member
}

// OrgMemberMiddleware verifies the user is a member of the org in the URL.
// Sets both org and member in context.
func OrgMemberMiddleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := auth.GetUserClaims(r.Context())
			if claims == nil {
				api.HandleError(w, api.ErrUnauthorized)
				return
			}

			orgIDStr := chi.URLParam(r, "orgID")
			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				api.HandleError(w, api.NewBadRequestError("invalid org ID"))
				return
			}

			org, err := svc.GetOrg(r.Context(), orgID)
			if err != nil {
				api.HandleError(w, api.ErrInternalServer)
				return
			}
			if org == nil {
				api.HandleError(w, api.NewNotFoundError("organization not found"))
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				api.HandleError(w, api.ErrUnauthorized)
				return
			}

			member, err := svc.GetMember(r.Context(), orgID, userID)
			if err != nil {
				api.HandleError(w, api.ErrInternalServer)
				return
			}
			if member == nil {
				api.HandleError(w, api.ErrForbidden)
				return
			}

			ctx := SetOrgInContext(r.Context(), org)
			ctx = SetMemberInContext(ctx, member)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OrgRoleMiddleware checks if the authenticated member has at least the required role.
// Must be used after OrgMemberMiddleware.
func OrgRoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			member := GetMemberFromContext(r.Context())
			if member == nil {
				api.HandleError(w, api.ErrForbidden)
				return
			}

			if !HasPermission(member.Role, requiredRole) {
				api.HandleError(w, api.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
