package generator

import (
	"strings"
	"testing"
)

func TestNameFormat(t *testing.T) {
	name := GenerateName(false)
	if !strings.Contains(name, "-") {
		t.Fatalf("Generated name does not contain an underscore")
	}
	if strings.ContainsAny(name, "0123456789") {
		t.Fatalf("Generated name contains numbers!")
	}
}

func TestNameRetries(t *testing.T) {
	name := GenerateName(true)
	if !strings.Contains(name, "-") {
		t.Fatalf("Generated name does not contain an underscore")
	}
	if !strings.ContainsAny(name, "0123456789") {
		t.Fatalf("Generated name doesn't contain a number")
	}

}

func TestDuplicateNames(t *testing.T) {
	firstName := GenerateName(true)
	secondName := GenerateName(true)
	if firstName == secondName {
		t.Fatalf("Duplicate generated names")
	}
}
