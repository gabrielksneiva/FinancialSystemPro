package finance

import "testing"

func TestAmountValidation(t *testing.T) {
	cases := []struct {
		v        int64
		currency string
		expectErr bool
	}{
		{v: 0, currency: "USD", expectErr: false},
		{v: 100, currency: "EUR", expectErr: false},
		{v: -1, currency: "USD", expectErr: true},
		{v: 10, currency: "usd", expectErr: true},
		{v: 10, currency: "US", expectErr: true},
	}
	for _, c := range cases {
		_, err := NewAmount(c.v, c.currency)
		if c.expectErr && err == nil {
													t.Errorf("expected error for value=%d currency=%s", c.v, c.currency)
		}
		if !c.expectErr && err != nil {
													t.Errorf("did not expect error for value=%d currency=%s: %v", c.v, c.currency, err)
		}
	}
}
