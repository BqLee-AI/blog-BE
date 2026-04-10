package handler

import "testing"

func TestRegisterRequestTokenPrefersRegistrationToken(t *testing.T) {
	req := registerRequest{
		RegistrationToken: "  new-token  ",
		Code:              "legacy-code",
	}

	if got := req.token(); got != "new-token" {
		t.Fatalf("expected registration_token to win, got %q", got)
	}
}

func TestRegisterRequestTokenFallsBackToCode(t *testing.T) {
	req := registerRequest{
		Code: "  legacy-code  ",
	}

	if got := req.token(); got != "legacy-code" {
		t.Fatalf("expected code fallback, got %q", got)
	}
}

func TestRegisterRequestTokenTrimsEmptyValues(t *testing.T) {
	req := registerRequest{}

	if got := req.token(); got != "" {
		t.Fatalf("expected empty token, got %q", got)
	}
}
