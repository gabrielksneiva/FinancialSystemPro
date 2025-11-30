package btc

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    bc "financial-system-pro/internal/blockchain"
    "github.com/shopspring/decimal"
)

func TestBtcConnector_BasicFlow(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req map[string]interface{}
        _ = json.NewDecoder(r.Body).Decode(&req)
        method := req["method"].(string)
        switch method {
        case "getaddressbalance":
            // return satoshi as string "100000000" (1 BTC)
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":"100000000"})
        case "sendrawtransaction":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":"txhash123"})
        case "getrawtransaction":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":map[string]interface{}{"confirmations":1}})
        case "getblockcount":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":2})
        case "getblockbyheight":
            _ = json.NewEncoder(w).Encode(map[string]interface{}{"result":map[string]interface{}{"hash":"blockhash"}})
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
        t.Fatalf("expected 1 BTC got %s", bal.String())
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
