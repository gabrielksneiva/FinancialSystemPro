package rpc

import (
    "context"
    "testing"

    "financial-system-pro/internal/blockchains/domain"
)

func TestSimulatedClient_BalanceAndTxLifecycle(t *testing.T) {
    c := NewSimulatedClient()
    ctx := context.Background()
    _ = c.Connect(ctx)

    amt, _ := domain.NewAmountFromString("5000")
    c.SetBalance("addr1", amt)
    b, err := c.GetBalance(ctx, "addr1")
    if err != nil {
        t.Fatalf("unexpected: %v", err)
    }
    if b.Value.String() != "5000" {
        t.Fatalf("expected 5000 got %s", b.Value.String())
    }

    tx := domain.Transaction{From: "addr1", To: "addr2", Amount: amt}
    hash, err := c.SubmitTransaction(ctx, tx)
    if err != nil {
        t.Fatalf("submit err: %v", err)
    }

    got, err := c.GetTransaction(ctx, hash)
    if err != nil {
        t.Fatalf("get tx err: %v", err)
    }
    if got.Status != domain.TxPending {
        t.Fatalf("expected pending")
    }

    c.ConfirmTransaction(hash, 10)
    got2, _ := c.GetTransaction(ctx, hash)
    if got2.Status != domain.TxConfirmed {
        t.Fatalf("expected confirmed")
    }
}

func TestSimulatedClient_BlockAdvanceAndMissingBalance(t *testing.T) {
    c := NewSimulatedClient()
    ctx := context.Background()
    c.AdvanceBlock(42)
    blk, err := c.GetLatestBlock(ctx)
    if err != nil {
        t.Fatalf("get latest block err: %v", err)
    }
    if blk.Number != 42 {
        t.Fatalf("expected block 42 got %d", blk.Number)
    }

    // missing balance should return ErrNotFound
    if _, err := c.GetBalance(ctx, "noaddr"); err == nil {
        t.Fatalf("expected not found")
    }
    // coverage: submit transaction with empty hash
    amt, _ := domain.NewAmountFromString("1")
    tx := domain.Transaction{From: "a", To: "b", Amount: amt}
    h, _ := c.SubmitTransaction(ctx, tx)
    if h == "" {
        t.Fatalf("expected generated hash")
    }
}
