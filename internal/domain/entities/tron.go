package entities

// TronWallet representa uma carteira Tron
type TronWallet struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key"`
}

// TronTransaction representa uma transação Tron
type TronTransaction struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TxHash      string `json:"tx_hash"`
	Status      string `json:"status"`
	Amount      int64  `json:"amount"`
	Timestamp   int64  `json:"timestamp"`
	Energy      int64  `json:"energy"`
}

// TronBalance representa o saldo de uma conta Tron
type TronBalance struct {
	Address    string `json:"address"`
	BalanceTRX string `json:"balance_trx"`
	BalanceSUN int64  `json:"balance_sun"`
}

// TronTestnetRequest representa uma requisição para operações na testnet
type TronTestnetRequest struct {
	Address    string `json:"address" validate:"required"`
	PrivateKey string `json:"private_key,omitempty"`
}

// TronTransactionRequest representa uma requisição para enviar transação
type TronTransactionRequest struct {
	FromAddress string `json:"from_address" validate:"required"`
	ToAddress   string `json:"to_address" validate:"required"`
	PrivateKey  string `json:"private_key" validate:"required"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
}
