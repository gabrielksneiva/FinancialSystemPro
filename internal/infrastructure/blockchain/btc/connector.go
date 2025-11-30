package btc

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "time"

    bc "financial-system-pro/internal/blockchain"
    "github.com/shopspring/decimal"
)

type BtcConnector struct {
    endpoint string
    httpc    *http.Client
}

func NewConnector(endpoint string) *BtcConnector {
    return &BtcConnector{endpoint: endpoint, httpc: http.DefaultClient}
}

// Implement Connector methods in a simplified way using a JSON-RPC style POST.
func (b *BtcConnector) FetchBalance(ctx context.Context, address bc.Address) (decimal.Decimal, error) {
    // For tests we expect the endpoint to reply with a numeric balance (satoshi) as string.
    type resp struct{ Result string `json:"result"` }
    var r resp
    if err := postJSONRPC(ctx, b.httpc, b.endpoint, "getaddressbalance", []interface{}{string(address)}, &r); err != nil {
        return decimal.Zero, err
    }
    // convert satoshi (string) to decimal BTC (8 decimals)
    d, err := decimal.NewFromString(r.Result)
    if err != nil {
        return decimal.Zero, err
    }
    return d.Shift(-8), nil
}

func (b *BtcConnector) SendTransaction(ctx context.Context, rawTx string) (bc.TxHash, error) {
    type resp struct{ Result string `json:"result"` }
    var r resp
    if err := postJSONRPC(ctx, b.httpc, b.endpoint, "sendrawtransaction", []interface{}{rawTx}, &r); err != nil {
        return "", err
    }
    return bc.TxHash(r.Result), nil
}

func (b *BtcConnector) GetTransactionStatus(ctx context.Context, hash bc.TxHash) (string, error) {
    type resp struct{ Result map[string]interface{} `json:"result"` }
    var r resp
    if err := postJSONRPC(ctx, b.httpc, b.endpoint, "getrawtransaction", []interface{}{string(hash), true}, &r); err != nil {
        return "", err
    }
    if r.Result == nil {
        return "pending", nil
    }
    if conf, ok := r.Result["confirmations"].(float64); ok && conf > 0 {
        return "confirmed", nil
    }
    return "pending", nil
}

func (b *BtcConnector) SyncLatestBlocks(ctx context.Context, since uint64) ([]bc.Block, error) {
    // simple implementation: ask for latest block height then fetch blocks
    type r1 struct{ Result uint64 `json:"result"` }
    var rr r1
    if err := postJSONRPC(ctx, b.httpc, b.endpoint, "getblockcount", nil, &rr); err != nil {
        return nil, err
    }
    latest := rr.Result
    out := []bc.Block{}
    for i := since + 1; i <= latest; i++ {
        var rb struct{ Result map[string]interface{} `json:"result"` }
        if err := postJSONRPC(ctx, b.httpc, b.endpoint, "getblockbyheight", []interface{}{i}, &rb); err != nil {
            return nil, err
        }
        h := ""
        if hh, ok := rb.Result["hash"].(string); ok {
            h = hh
        }
        out = append(out, bc.Block{Number: i, Hash: h, Time: time.Now()})
    }
    return out, nil
}

// postJSONRPC performs a simple JSON-RPC call and decodes into out.
func postJSONRPC(ctx context.Context, httpc *http.Client, endpoint, method string, params []interface{}, out interface{}) error {
    reqObj := map[string]interface{}{"jsonrpc": "1.0", "id": "1", "method": method, "params": params}
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

