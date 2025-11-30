package eth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bc "financial-system-pro/internal/blockchain"

	"github.com/shopspring/decimal"
)

func TestEthConnector_BasicFlow(t *testing.T) {
	// mock RPC server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		method := req["method"].(string)
		switch method {
		case "eth_getBalance":
			// return 0xde0b6b3a7640000 (1 ETH)
			resp := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0xde0b6b3a7640000"}
			_ = json.NewEncoder(w).Encode(resp)
		case "eth_sendRawTransaction":
			resp := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0xdeadbeef"}
			_ = json.NewEncoder(w).Encode(resp)
		case "eth_getTransactionReceipt":
			// simulate confirmed
			resp := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": map[string]interface{}{"status": "0x1"}}
			_ = json.NewEncoder(w).Encode(resp)
		case "eth_blockNumber":
			resp := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0x10"}
			_ = json.NewEncoder(w).Encode(resp)
		case "eth_getBlockByNumber":
			// return basic block
			resp := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": map[string]interface{}{"number": "0x10", "hash": "0xabc"}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "error": map[string]interface{}{"code": -1, "message": "unknown method"}})
		}
	}))
	defer srv.Close()

	conn := NewConnector(srv.URL)
	// balance
	bal, err := conn.FetchBalance(context.Background(), bc.Address("0xabc"))
	if err != nil {
		t.Fatalf("FetchBalance err: %v", err)
	}
	if !bal.Equal(decimal.NewFromFloat(1.0)) {
		t.Fatalf("expected 1 ETH got %s", bal.String())
	}
	// send tx
	txh, err := conn.SendTransaction(context.Background(), "0xfake")
	if err != nil || txh == "" {
		t.Fatalf("SendTransaction err: %v txh:%s", err, txh)
	}
	// status
	st, err := conn.GetTransactionStatus(context.Background(), txh)
	if err != nil || st != "confirmed" {
		t.Fatalf("GetTransactionStatus err:%v status:%s", err, st)
	}
	// sync blocks from 0x0
	blks, err := conn.SyncLatestBlocks(context.Background(), 0)
	if err != nil {
		t.Fatalf("SyncLatestBlocks err: %v", err)
	}
	if len(blks) == 0 {
		t.Fatalf("expected blocks, got none")
	}
}
