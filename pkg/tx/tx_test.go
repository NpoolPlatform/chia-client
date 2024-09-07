package tx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnsignedTx(t *testing.T) {
	spends, err := GenerateUnsignedTransaction("90", "0x87d86f291b7da8b2a4e197ffaa4a2d05ef5cbbc5aafd37829f4cf2147c6ec915", "11")
	assert.Nil(t, err)
	for _, spend := range spends {
		fmt.Println("coin: ", spend.Coin)
		fmt.Println("solution: ", spend.Solution)
		fmt.Println("puzzle: ", spend.PuzzleReveal)
	}
}
