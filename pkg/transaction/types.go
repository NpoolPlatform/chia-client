package transaction

import "github.com/chia-network/go-chia-libs/pkg/types"

const (
	aggsig_data = "ccd5bb71183532bff220ba46c268991a3ff07eb358e8255a65c30a2dce0e5fbb"
)

type UnsignedTx struct {
	From         string
	SpentCoinIDs []string
	Spends       []*UnsignedSpend
}

type UnsignedSpend struct {
	SpentCoinIDs []string
	Coin         *types.Coin
	Solution     []byte
	Message      string
}
