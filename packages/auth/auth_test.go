package auth

import (
	"strings"
	"testing"
)

func TestLoginClientRequiresEnvironmentToken(t *testing.T) {
	t.Setenv("WHOP_API_TOKEN", "")
	_, err := LoginClient("example-license")
	if err == nil || !strings.Contains(err.Error(), "WHOP_API_TOKEN") {
		t.Fatalf("expected a missing token error, got %v", err)
	}
}
