package admin_test

import (
	"os"
	"testing"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/admin"
)

func TestIsAdmin_DefaultUsername(t *testing.T) {
	os.Unsetenv("ADMIN_USERNAMES")
	if !admin.IsAdmin("admin") {
		t.Fatal("expected admin user to be admin")
	}
	if admin.IsAdmin("regular") {
		t.Fatal("expected regular user to not be admin")
	}
}

func TestIsAdmin_EnvOverride(t *testing.T) {
	t.Setenv("ADMIN_USERNAMES", "ops,lead")
	if !admin.IsAdmin("OPS") {
		t.Fatal("expected case-insensitive match")
	}
	if admin.IsAdmin("admin") {
		t.Fatal("default admin should not apply when env is set")
	}
}
