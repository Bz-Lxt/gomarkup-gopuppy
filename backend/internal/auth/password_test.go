package auth

import "testing"

func TestHashCostAndCompare(t *testing.T) {
	h, err := HashPassword("Puppy123!")
	if err != nil {
		t.Fatal(err)
	}
	if h == "Puppy123!" {
		t.Fatal("plaintext stored")
	}
	if err := ComparePassword(h, "Puppy123!"); err != nil {
		t.Fatal(err)
	}
	if err := ComparePassword(h, "wrong"); err == nil {
		t.Fatal("should reject")
	}
}

func TestShortPasswordRejected(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("expected error")
	}
}
