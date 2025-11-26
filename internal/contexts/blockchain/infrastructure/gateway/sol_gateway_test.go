package gateway

import (
	"context"
	"testing"
)

func TestSOLGateway_Basic(t *testing.T) {
	g := NewSOLGatewayFromEnv()
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
	h, err := g.Broadcast(context.Background(), w.Address, w.Address, 123, w.PrivateKey)
	if err != nil || len(h) == 0 {
		t.Fatalf("broadcast failed: %v", err)
	}
	st, err := g.GetStatus(context.Background(), h)
	if err != nil || st.Status == "" {
		t.Fatalf("status failed: %v", err)
	}
	bal, err := g.GetBalance(context.Background(), w.Address)
	if err != nil || bal <= 0 {
		t.Fatalf("balance failed: %v bal=%d", err, bal)
	}
}
