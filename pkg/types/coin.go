package types

import (
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type CoinSpend struct {
	Coin         types1.Coin `json:"coin"`
	PuzzleReveal string      `json:"puzzle_reveal"`
	Solution     string      `json:"solution"`
}
