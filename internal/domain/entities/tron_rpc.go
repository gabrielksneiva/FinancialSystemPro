package entities

// TronRPCAccount representa uma conta Tron via RPC
type TronRPCAccount struct {
	Address  string `json:"address"`
	CodeHash string `json:"codeHash"`
	Balance  int64  `json:"balance"`
	Nonce    int64  `json:"nonce"`
}

// TronRPCTransaction representa uma transação Tron via RPC
type TronRPCTransaction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Data     string `json:"data"`
	Nonce    string `json:"nonce"`
	Hash     string `json:"hash"`
}

// TronRPCBlock representa um bloco Tron via RPC
type TronRPCBlock struct {
	Number       string   `json:"number"`
	Hash         string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	Timestamp    string   `json:"timestamp"`
	Miner        string   `json:"miner"`
	Difficulty   string   `json:"difficulty"`
	GasLimit     string   `json:"gasLimit"`
	GasUsed      string   `json:"gasUsed"`
	Transactions []string `json:"transactions"`
}

// TronRPCReceipt representa um recibo de transação Tron via RPC
type TronRPCReceipt struct {
	TransactionHash string        `json:"transactionHash"`
	BlockNumber     string        `json:"blockNumber"`
	BlockHash       string        `json:"blockHash"`
	From            string        `json:"from"`
	To              string        `json:"to"`
	Status          string        `json:"status"`
	GasUsed         string        `json:"gasUsed"`
	Logs            []interface{} `json:"logs"`
	Confirmations   int64         `json:"confirmations"`
}

// TronNetworkStatus representa o status da rede Tron
type TronNetworkStatus struct {
	NetworkVersion string  `json:"network_version"`
	LatestBlock    int64   `json:"latest_block"`
	Timestamp      int64   `json:"timestamp"`
	ResponseTime   float64 `json:"response_time_ms"`
	IsConnected    bool    `json:"is_connected"`
}

// TronGasEstimate representa uma estimativa de gas/energia
type TronGasEstimate struct {
	EstimatedCost string `json:"estimated_cost"`
	StandardGas   int64  `json:"standard_gas"`
	FastGas       int64  `json:"fast_gas"`
	InstantGas    int64  `json:"instant_gas"`
	GasPrice      int64  `json:"gas_price"`
}

// TronContractCall representa uma chamada de contrato
type TronContractCall struct {
	To    string `json:"to"`
	From  string `json:"from"`
	Data  string `json:"data"`
	Value string `json:"value,omitempty"`
	Gas   string `json:"gas,omitempty"`
}

// TronRPCMethod lista os métodos RPC disponíveis
type TronRPCMethod struct {
	Name        string
	Description string
	Returns     string
	Params      []string
}

// GetAvailableMethods retorna os métodos RPC disponíveis para Tron
func GetAvailableTronRPCMethods() []TronRPCMethod {
	return []TronRPCMethod{
		{
			Name:        "eth_blockNumber",
			Description: "Retorna o número do bloco mais recente",
			Params:      []string{},
			Returns:     "QUANTITY - número do bloco",
		},
		{
			Name:        "eth_getBalance",
			Description: "Retorna o saldo da conta Tron",
			Params:      []string{"DATA - endereço Tron", "QUANTITY|TAG - bloco"},
			Returns:     "QUANTITY - saldo em SUN",
		},
		{
			Name:        "eth_sendRawTransaction",
			Description: "Envia uma transação assinada",
			Params:      []string{"DATA - transação serializada"},
			Returns:     "DATA - hash da transação",
		},
		{
			Name:        "eth_getTransactionByHash",
			Description: "Retorna informações da transação",
			Params:      []string{"DATA - hash da transação"},
			Returns:     "OBJECT - objeto da transação",
		},
		{
			Name:        "eth_getTransactionReceipt",
			Description: "Retorna o recibo da transação",
			Params:      []string{"DATA - hash da transação"},
			Returns:     "OBJECT - objeto do recibo ou null",
		},
		{
			Name:        "eth_estimateGas",
			Description: "Estima o gas necessário",
			Params:      []string{"OBJECT - objeto da transação"},
			Returns:     "QUANTITY - gas estimado",
		},
		{
			Name:        "eth_gasPrice",
			Description: "Retorna o preço atual do gas",
			Params:      []string{},
			Returns:     "QUANTITY - preço do gas em SUN",
		},
		{
			Name:        "eth_call",
			Description: "Executa uma chamada de contrato",
			Params:      []string{"OBJECT - objeto da transação", "QUANTITY|TAG - bloco"},
			Returns:     "DATA - resultado da chamada",
		},
		{
			Name:        "eth_getCode",
			Description: "Retorna o código do contrato",
			Params:      []string{"DATA - endereço", "QUANTITY|TAG - bloco"},
			Returns:     "DATA - bytecode do contrato",
		},
		{
			Name:        "eth_getTransactionCount",
			Description: "Retorna o número de transações",
			Params:      []string{"DATA - endereço", "QUANTITY|TAG - bloco"},
			Returns:     "QUANTITY - número de transações",
		},
	}
}
