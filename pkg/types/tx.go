package types

import (
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type Payment struct {
	PuzzleHash types1.Bytes32 `json:"puzzle_hash"`
	Amount     string         `json:"amount"`
	Memos      []string       `json:"memos,omitempty"`
}

type CoinSpend struct {
	Coin         types1.Coin `json:"coin"`
	PuzzleReveal string      `json:"puzzle_reveal"`
	Solution     string      `json:"solution"`
}

type UnsignedTx struct {
	CoinSpends []*CoinSpend `json:"coin_spends"`
	Messages   []string     `json:"messages"`
}
