package gateway

import (
	"context"
	"testing"
)

func TestBTCGateway_Basic(t *testing.T) {
	g := NewBTCGatewayFromEnv()
	w, err := g.GenerateWallet(context.Background())
	if err != nil {
		t.Fatalf("generate wallet: %v", err)
	}
	if !g.ValidateAddress(w.Address) {
		t.Fatalf("invalid generated address: %s", w.Address)
	}
	fq, err := g.EstimateFee(context.Background(), w.Address, w.Address, 1)
	if err != nil || fq.EstimatedFee <= 0 {
		t.Fatalf("fee estimate invalid: %+v, err=%v", fq, err)
	}
	h, err := g.Broadcast(context.Background(), w.Address, w.Address, 123, "priv")
	if err != nil || len(h) == 0 {
		t.Fatalf("broadcast failed: %v", err)
	}
	st, err := g.GetStatus(context.Background(), h)
	if err != nil || st.Status == "" {
		t.Fatalf("status failed: %v", err)
	}
	bal, _ := g.GetBalance(context.Background(), w.Address)
	if bal == 0 {
		t.Fatalf("expected non-zero default balance")
	}
}
