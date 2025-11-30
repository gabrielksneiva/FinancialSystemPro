package eth

import (
    "context"
    "encoding/hex"
    "encoding/json"
    "math/big"
    "time"

    bc "financial-system-pro/internal/blockchain"

    "github.com/shopspring/decimal"
)

type EthConnector struct {
    client *Client
}

func NewConnector(endpoint string) *EthConnector {
    return &EthConnector{client: NewClient(endpoint)}
}

func hexToBigInt(hexstr string) (*big.Int, error) {
    // strip 0x
    if len(hexstr) >= 2 && hexstr[0:2] == "0x" {
        hexstr = hexstr[2:]
    }
    if hexstr == "" {
        return big.NewInt(0), nil
    }
    // pad leading zero if odd length
    if len(hexstr)%2 == 1 {
        hexstr = "0" + hexstr
    }
    b, err := hex.DecodeString(hexstr)
    if err != nil {
        return nil, err
    }
    i := new(big.Int).SetBytes(b)
    return i, nil
}

func (e *EthConnector) FetchBalance(ctx context.Context, address bc.Address) (decimal.Decimal, error) {
    res, err := e.client.call(ctx, "eth_getBalance", []interface{}{string(address), "latest"})
    if err != nil {
        return decimal.Zero, err
    }
    var hexstr string
    if err := json.Unmarshal(res, &hexstr); err != nil {
        return decimal.Zero, err
    }
    bi, err := hexToBigInt(hexstr)
    if err != nil {
        return decimal.Zero, err
    }
    // convert wei to ether (18 decimals)
    d := decimal.NewFromBigInt(bi, -18)
    return d, nil
}

func (e *EthConnector) SendTransaction(ctx context.Context, rawTx string) (bc.TxHash, error) {
    res, err := e.client.call(ctx, "eth_sendRawTransaction", []interface{}{rawTx})
    if err != nil {
        return "", err
    }
    var txh string
    if err := json.Unmarshal(res, &txh); err != nil {
        return "", err
    }
    return bc.TxHash(txh), nil
}

func (e *EthConnector) GetTransactionStatus(ctx context.Context, hash bc.TxHash) (string, error) {
    res, err := e.client.call(ctx, "eth_getTransactionReceipt", []interface{}{string(hash)})
    if err != nil {
        return "", err
    }
    if string(res) == "null" {
        return "pending", nil
    }
    var obj map[string]interface{}
    if err := json.Unmarshal(res, &obj); err != nil {
        return "", err
    }
    if s, ok := obj["status"]; ok {
        if s == "0x1" || s == float64(1) {
            return "confirmed", nil
        }
        return "failed", nil
    }
    return "unknown", nil
}

func (e *EthConnector) SyncLatestBlocks(ctx context.Context, since uint64) ([]bc.Block, error) {
    // get latest block number
    res, err := e.client.call(ctx, "eth_blockNumber", []interface{}{})
    if err != nil {
        return nil, err
    }
    var hexstr string
    if err := json.Unmarshal(res, &hexstr); err != nil {
        return nil, err
    }
    bi, err := hexToBigInt(hexstr)
    if err != nil {
        return nil, err
    }
    latest := bi.Uint64()
    out := make([]bc.Block, 0)
    for i := since + 1; i <= latest; i++ {
        // format hex
        h := "0x" + big.NewInt(0).SetUint64(i).Text(16)
        r, err := e.client.call(ctx, "eth_getBlockByNumber", []interface{}{h, false})
        if err != nil {
            return nil, err
        }
        if string(r) == "null" {
            continue
        }
        var obj map[string]interface{}
        if err := json.Unmarshal(r, &obj); err != nil {
            return nil, err
        }
        bnum := uint64(0)
        if n, ok := obj["number"].(string); ok {
            nb, _ := hexToBigInt(n)
            bnum = nb.Uint64()
        }
        bh := ""
        if hsh, ok := obj["hash"].(string); ok {
            bh = hsh
        }
        out = append(out, bc.Block{Number: bnum, Hash: bh, Time: time.Now()})
    }
    return out, nil
}
