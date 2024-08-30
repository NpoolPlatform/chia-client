package types

type Payment struct {
	PuzzleHash Bytes32 `json:"puzzle_hash" streamable:""`
	Amount     uint64  `json:"amount"`
	Memos      []Bytes `json:"memos,omitempty"`
}
