package uploads

import (
	"path/filepath"
	"testing"
)

func TestManagedRelativePath(t *testing.T) {
	t.Parallel()

	relative, ok := ManagedRelativePath("https://example.com/api/uploads/profile-avatars/u/abc.png", "/api/uploads/")
	if !ok {
		t.Fatalf("expected managed path to be detected")
	}
	if relative != "profile-avatars/u/abc.png" {
		t.Fatalf("relative path = %q, want %q", relative, "profile-avatars/u/abc.png")
	}

	if _, ok := ManagedRelativePath("https://example.com/other/path.png", "/api/uploads/"); ok {
		t.Fatalf("unexpected managed path detection for external URL")
	}
}

func TestResolvePathPreventsTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target, err := ResolvePath(root, "../outside/../../avatar.png")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}

	expected := filepath.Join(root, "avatar.png")
	if target != expected {
		t.Fatalf("target = %q, want %q", target, expected)
	}
}
