package types

type Payment struct {
	PuzzleHash string   `json:"puzzle_hash" streamable:""`
	Amount     string   `json:"amount"`
	Memos      []string `json:"memos,omitempty"`
}
