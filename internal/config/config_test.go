package config

import (
	"testing"
	"strings"
)

func TestValidateSSLMode_Allowed(t *testing.T) {
	allowed := []string{
		"disable",
		"allow",
		"prefer",
		"require",
		"verify-ca",
		"verify-full",
		"prefer ",
                " DISABLE",
	}

	for _, v := range allowed {
		t.Run(v, func(t *testing.T) {
			got, err := validateSSLMode(v)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got != strings.ToLower(strings.TrimSpace(v)) {
				t.Fatalf("expected %q, got %q", v, got)
			}
		})
	}
}

func TestValidateSSLMode_Rejected(t *testing.T) {
	rejected := []string{
		"",
		"0",
		"1",
		"enabled",
		"verify",
		"verifyfull",
		"prefers ",
		" DlSABLE",
	}

	for _, v := range rejected {
		t.Run(v, func(t *testing.T) {
			_, err := validateSSLMode(v)
			if err == nil {
				t.Fatalf("expected error for %q", v)
			}
		})
	}
}
