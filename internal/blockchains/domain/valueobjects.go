package domain

import (
	"errors"
	"math/big"
	"regexp"
	"time"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
	ErrInvalidAmount  = errors.New("invalid amount")
)

type Address string

func NewAddress(s string) (Address, error) {
	if s == "" {
		return "", ErrInvalidAddress
	}
	re := regexp.MustCompile(`^[0-9a-zA-Z:._-]{5,120}$`)
	if !re.MatchString(s) {
		return "", ErrInvalidAddress
	}
	return Address(s), nil
}

type Amount struct {
	Value *big.Int
}

func NewAmountFromString(s string) (Amount, error) {
	if s == "" {
		return Amount{}, ErrInvalidAmount
	}
	i := new(big.Int)
	_, ok := i.SetString(s, 10)
	if !ok {
		return Amount{}, ErrInvalidAmount
	}
	if i.Sign() < 0 {
		return Amount{}, ErrInvalidAmount
	}
	return Amount{Value: i}, nil
}

type TxHash string

func NewTxHash(h string) (TxHash, error) {
	if h == "" {
		return "", errors.New("empty tx hash")
	}
	return TxHash(h), nil
}

type Block struct {
	Number uint64
	Hash   string
	Time   time.Time
}

type TxStatus string

const (
	TxPending   TxStatus = "PENDING"
	TxConfirmed TxStatus = "CONFIRMED"
	TxFailed    TxStatus = "FAILED"
)

type Transaction struct {
	Hash      string
	From      Address
	To        Address
	Amount    Amount
	Status    TxStatus
	BlockNum  uint64
	CreatedAt time.Time
}
