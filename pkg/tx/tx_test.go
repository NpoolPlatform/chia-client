package tx

import (
	"fmt"
	"testing"

	types1 "github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUnsignedTransaction(t *testing.T) {
	parentCoinID1, err := types1.Bytes32FromHexString("0x3e74e639645d5efa616a39edcdc92b8a56d027f76b97e93d8169e8e1f9dd6226")
	assert.Nil(t, err)
	puzzleHash, err := types1.Bytes32FromHexString("0x09f0a008ea9756842e235660cf3bc8c29bf7486f0904d4912b9ba93781aa08f8")
	assert.Nil(t, err)
	coins := []*types1.Coin{
		{
			ParentCoinInfo: parentCoinID1,
			PuzzleHash:     puzzleHash,
			Amount:         2000,
		},
	}
	fmt.Println("id: ", coins[0].ID().String())
	from := "txch1eusf0gslhc6s7xv2w95t5rylf4ftv4r0wq4t0r8p82zfmuq2nzfscn3xfe"
	to := "txch1jlk3xekckp9uftk7vlec69ludumgrjfhlwgehvmst596q2t2qhjsarzrvh"
	amount := "1000"
	fee := "15"
	additionalData := "ccd5bb71183532bff220ba46c268991a3ff07eb358e8255a65c30a2dce0e5fbb"
	unsignedTx, err := GenerateUnsignedTransaction(from, to, amount, fee, coins, additionalData)
	assert.Nil(t, err)
	fmt.Println("unsignedTx: messages ", unsignedTx.Messages)
	fmt.Println("unsignedTx: solution ", unsignedTx.CoinSpends[0].Solution)
	fmt.Println("unsignedTx: puzzle ", unsignedTx.CoinSpends[0].PuzzleReveal)
}
