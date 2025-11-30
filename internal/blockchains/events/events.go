package events

import "time"

type EventType string

const (
	EventNewBlock           EventType = "NewBlock"
	EventNewTransaction     EventType = "NewTransaction"
	EventTransactionConfirm EventType = "TransactionConfirmed"
)

type NewBlockEvent struct {
	Chain       string
	BlockNumber uint64
	BlockHash   string
	Timestamp   time.Time
}

type NewTransactionEvent struct {
	Chain  string
	TxHash string
}

type TransactionConfirmedEvent struct {
	Chain       string
	TxHash      string
	BlockNumber uint64
}
