package types

type CoinSpend struct {
	Coin         Coin   `json:"coin"`
	PuzzleReveal string `json:"puzzle_reveal"`
	Solution     string `json:"solution"`
}

