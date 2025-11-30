package sol

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    bc "financial-system-pro/internal/blockchain"
    "github.com/shopspring/decimal"
)

type SolConnector struct{
    endpoint string
    httpc    *http.Client
}

func NewConnector(endpoint string) *SolConnector {
    return &SolConnector{endpoint: endpoint, httpc: http.DefaultClient}
}

func (s *SolConnector) FetchBalance(ctx context.Context, address bc.Address) (decimal.Decimal, error) {
    // solana getBalance returns lamports (u64). We'll expect a JSON-RPC response.
    var out struct{ Result struct{ Value uint64 `json:"value"` } `json:"result"` }
    if err := postJSONRPC(ctx, s.httpc, s.endpoint, "getBalance", []interface{}{string(address)}, &out); err != nil {
        return decimal.Zero, err
    }
    // convert lamports to SOL (9 decimals)
    d := decimal.NewFromInt(int64(out.Result.Value)).Shift(-9)
    return d, nil
}

func (s *SolConnector) SendTransaction(ctx context.Context, rawTx string) (bc.TxHash, error) {
    var out struct{ Result string `json:"result"` }
    if err := postJSONRPC(ctx, s.httpc, s.endpoint, "sendTransaction", []interface{}{rawTx}, &out); err != nil {
        return "", err
    }
    return bc.TxHash(out.Result), nil
}

func (s *SolConnector) GetTransactionStatus(ctx context.Context, hash bc.TxHash) (string, error) {
    var out struct{ Result interface{} `json:"result"` }
    if err := postJSONRPC(ctx, s.httpc, s.endpoint, "getConfirmedTransaction", []interface{}{string(hash)}, &out); err != nil {
        return "", err
    }
    if out.Result == nil {
        return "pending", nil
    }
    return "confirmed", nil
}

func (s *SolConnector) SyncLatestBlocks(ctx context.Context, since uint64) ([]bc.Block, error) {
    var r struct{ Result uint64 `json:"result"` }
    if err := postJSONRPC(ctx, s.httpc, s.endpoint, "getSlot", nil, &r); err != nil {
        return nil, err
    }
    latest := r.Result
    out := []bc.Block{}
    for i := since + 1; i <= latest; i++ {
        // Solana doesn't expose block by number in the same way; simulate small block info
        out = append(out, bc.Block{Number: i, Hash: fmt.Sprintf("slot-%d", i), Time: time.Now()})
    }
    return out, nil
}

// helper
func postJSONRPC(ctx context.Context, httpc *http.Client, endpoint, method string, params []interface{}, out interface{}) error {
    reqObj := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
    b, _ := json.Marshal(reqObj)
    req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := httpc.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return json.NewDecoder(resp.Body).Decode(out)
}

