package tools

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreatePermissionRequest_Fields(t *testing.T) {
	toolID := uuid.New()
	req := CreatePermissionRequest{
		ToolID:      toolID,
		ToolSource:  "system",
		GranteeType: "agent",
		GranteeID:   uuid.New().String(),
		Permission:  "execute",
	}
	assert.Equal(t, toolID, req.ToolID)
	assert.Equal(t, "system", req.ToolSource)
	assert.Equal(t, "agent", req.GranteeType)
	assert.Equal(t, "execute", req.Permission)
}

func TestToolPermission_Struct(t *testing.T) {
	perm := ToolPermission{
		ID:          uuid.New(),
		ToolID:      uuid.New(),
		ToolSource:  "agent",
		GranteeType: "role",
		GranteeID:   "admin",
		Permission:  "configure",
		GrantedBy:   uuid.New(),
	}
	assert.Equal(t, "agent", perm.ToolSource)
	assert.Equal(t, "role", perm.GranteeType)
	assert.Equal(t, "admin", perm.GranteeID)
	assert.Equal(t, "configure", perm.Permission)
}

func TestPermissionTypes(t *testing.T) {
	tests := []struct {
		name       string
		permission string
	}{
		{"execute permission", "execute"},
		{"configure permission", "configure"},
		{"admin permission", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreatePermissionRequest{
				ToolID:      uuid.New(),
				ToolSource:  "system",
				GranteeType: "agent",
				GranteeID:   uuid.New().String(),
				Permission:  tt.permission,
			}
			assert.Equal(t, tt.permission, req.Permission)
		})
	}
}

func TestGranteeTypes(t *testing.T) {
	tests := []struct {
		name        string
		granteeType string
	}{
		{"agent grantee", "agent"},
		{"role grantee", "role"},
		{"org grantee", "org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreatePermissionRequest{
				ToolID:      uuid.New(),
				ToolSource:  "system",
				GranteeType: tt.granteeType,
				GranteeID:   uuid.New().String(),
				Permission:  "execute",
			}
			assert.Equal(t, tt.granteeType, req.GranteeType)
		})
	}
}

func TestToolSources(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"system source", "system"},
		{"agent source", "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := ToolPermission{
				ID:         uuid.New(),
				ToolID:     uuid.New(),
				ToolSource: tt.source,
			}
			assert.Equal(t, tt.source, perm.ToolSource)
		})
	}
}
