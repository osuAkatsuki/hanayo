package user

import "testing"

func TestValidateUsernameAllowsBadGirl(t *testing.T) {
	if validationError := ValidateUsername("Bad Girl"); validationError != "" {
		t.Fatalf("expected Bad Girl to be allowed, got %q", validationError)
	}
}
