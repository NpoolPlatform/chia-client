package rpc

import (
	"fmt"
	"testing"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGetAggSigAdditionalData(t *testing.T) {
	client, err := GetFullNodeClient()
	assert.Nil(t, err)
	data, err := client.GetAggSigAdditionalData()
	assert.Nil(t, err)
	fmt.Println("data: ", *data)
	from, _ := types1.Bytes32FromHexString("0x5baecec1bc13676097ad96b29a73d599d93871b75981ec973f5c376486d08b4d")
	coins, err := client.GetCoinsByPuzzleHash(&rpc.GetCoinRecordsByPuzzleHashOptions{
		IncludeSpentCoins: false,
		PuzzleHash:        from,
	})
	assert.Nil(t, err)
	fmt.Println("coins: ", coins)
}
