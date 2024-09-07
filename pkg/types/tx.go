package types

import (
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type Payment struct {
	PuzzleHash types1.Bytes32 `json:"puzzle_hash" streamable:""`
	Amount     string         `json:"amount"`
	Memos      []string       `json:"memos,omitempty"`
}
