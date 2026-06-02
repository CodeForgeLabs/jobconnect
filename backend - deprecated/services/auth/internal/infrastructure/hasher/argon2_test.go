package hasher

import "testing"

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	h := NewArgon2Hasher()
	hash, err := h.Hash("Password1!")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	ok, err := h.Verify("Password1!", hash)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}
	// Wrong password should not verify.
	ok, err = h.Verify("WrongPassword", hash)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if ok {
		t.Fatalf("expected wrong password not to verify")
	}
}
