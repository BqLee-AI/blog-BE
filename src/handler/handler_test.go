package handler

import "testing"

func TestRegisterRequestTokenPrefersRegistrationToken(t *testing.T) {
	req := registerRequest{
		RegistrationToken: "  new-token  ",
		Code:              "legacy-code",
	}

	got, useRegistrationToken := req.verificationCredential()
	if got != "new-token" {
		t.Fatalf("expected registration_token to win, got %q", got)
	}
	if !useRegistrationToken {
		t.Fatal("expected registration_token flow to be selected")
	}
}

func TestRegisterRequestTokenFallsBackToCode(t *testing.T) {
	req := registerRequest{
		Code: "  legacy-code  ",
	}

	got, useRegistrationToken := req.verificationCredential()
	if got != "legacy-code" {
		t.Fatalf("expected code fallback, got %q", got)
	}
	if useRegistrationToken {
		t.Fatal("expected legacy code flow to be selected")
	}
}

func TestRegisterRequestTokenTrimsEmptyValues(t *testing.T) {
	req := registerRequest{}

	got, useRegistrationToken := req.verificationCredential()
	if got != "" {
		t.Fatalf("expected empty token, got %q", got)
	}
	if useRegistrationToken {
		t.Fatal("expected empty request to use legacy flow flag")
	}
}
