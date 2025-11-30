package eth

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type rpcRequest struct {
    JSONRPC string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params  []interface{} `json:"params"`
    ID      int           `json:"id"`
}

type rpcResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Result  json.RawMessage `json:"result"`
    Error   *rpcError       `json:"error"`
}

type rpcError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// Client is a minimal JSON-RPC client for Ethereum-like nodes.
type Client struct {
    endpoint string
    httpc    *http.Client
}

func NewClient(endpoint string) *Client {
    return &Client{endpoint: endpoint, httpc: http.DefaultClient}
}

func (c *Client) call(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
    reqObj := rpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1}
    b, _ := json.Marshal(reqObj)
    req, _ := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.httpc.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    var r rpcResponse
    if err := json.Unmarshal(body, &r); err != nil {
        return nil, err
    }
    if r.Error != nil {
        return nil, fmt.Errorf("rpc error %d: %s", r.Error.Code, r.Error.Message)
    }
    return r.Result, nil
}
