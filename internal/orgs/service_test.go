package orgs

import "testing"

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name     string
		member   string
		required string
		want     bool
	}{
		{"owner can do anything", RoleOwner, RoleOwner, true},
		{"owner satisfies admin", RoleOwner, RoleAdmin, true},
		{"admin satisfies editor", RoleAdmin, RoleEditor, true},
		{"admin satisfies viewer", RoleAdmin, RoleViewer, true},
		{"editor cannot admin", RoleEditor, RoleAdmin, false},
		{"viewer cannot edit", RoleViewer, RoleEditor, false},
		{"viewer is viewer", RoleViewer, RoleViewer, true},
		{"unknown role fails", "guest", RoleViewer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(tt.member, tt.required)
			if got != tt.want {
				t.Errorf("HasPermission(%q, %q) = %v, want %v", tt.member, tt.required, got, tt.want)
			}
		})
	}
}

func TestRoleLevel(t *testing.T) {
	if roleLevel(RoleOwner) <= roleLevel(RoleAdmin) {
		t.Error("owner should be higher than admin")
	}
	if roleLevel(RoleAdmin) <= roleLevel(RoleEditor) {
		t.Error("admin should be higher than editor")
	}
	if roleLevel(RoleEditor) <= roleLevel(RoleViewer) {
		t.Error("editor should be higher than viewer")
	}
	if roleLevel("unknown") != 0 {
		t.Error("unknown role should return 0")
	}
}
