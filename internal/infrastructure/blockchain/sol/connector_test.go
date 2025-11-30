package sol

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    bc "financial-system-pro/internal/blockchain"
    "github.com/shopspring/decimal"
)

func TestSolConnector_BasicFlow(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req map[string]interface{}
        _ = json.NewDecoder(r.Body).Decode(&req)
        method := req["method"].(string)
        switch method {
        case "getBalance":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":map[string]interface{}{"value":1000000000}}) // 1 SOL in lamports
        case "sendTransaction":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":"txhashsol"})
        case "getConfirmedTransaction":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":map[string]interface{}{"meta":"ok"}})
        case "getSlot":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":2})
        default:
            w.WriteHeader(500)
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"error":"unknown method"})
        }
    }))
    defer srv.Close()

    c := NewConnector(srv.URL)
    bal, err := c.FetchBalance(context.Background(), bc.Address("addr1"))
    if err != nil {
        t.Fatalf("FetchBalance err: %v", err)
    }
    if !bal.Equal(decimal.NewFromFloat(1.0)) {
        t.Fatalf("expected 1 SOL got %s", bal.String())
    }
    txh, err := c.SendTransaction(context.Background(), "rawtx")
    if err != nil || txh == "" {
        t.Fatalf("SendTransaction err:%v txh:%s", err, txh)
    }
    st, err := c.GetTransactionStatus(context.Background(), txh)
    if err != nil || st != "confirmed" {
        t.Fatalf("GetTransactionStatus err:%v status:%s", err, st)
    }
    blks, err := c.SyncLatestBlocks(context.Background(), 0)
    if err != nil || len(blks) == 0 {
        t.Fatalf("SyncLatestBlocks err:%v len=%d", err, len(blks))
    }
}
