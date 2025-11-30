package domain

import "testing"

func TestNewAddressValidation(t *testing.T) {
    // too short
    if _, err := NewAddress("ab"); err == nil {
        t.Fatalf("expected invalid for short")
    }
    // invalid chars
    if _, err := NewAddress("invalid!addr"); err == nil {
        t.Fatalf("expected invalid for chars")
    }
    // valid
    if a, err := NewAddress("valid_addr-01"); err != nil || a == "" {
        t.Fatalf("expected valid")
    }
}

func TestNewAddress_LengthBounds(t *testing.T) {
    // create string of length 121 should be invalid (max 120 in regex)
    long := make([]byte, 121)
    for i := range long {
        long[i] = 'a'
    }
    if _, err := NewAddress(string(long)); err == nil {
        t.Fatalf("expected invalid for too long address")
    }
}

func TestNewAddress_InvalidCharacters(t *testing.T) {
    // spaces and punctuation not allowed
    if _, err := NewAddress("has space"); err == nil {
        t.Fatalf("expected invalid for space")
    }
    if _, err := NewAddress("semi;colon"); err == nil {
        t.Fatalf("expected invalid for semicolon")
    }
}
