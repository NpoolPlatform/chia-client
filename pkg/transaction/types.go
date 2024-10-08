package transaction

import "github.com/chia-network/go-chia-libs/pkg/types"

type UnsignedTx struct {
	From         string
	SpentCoinIDs []string
	Spends       []*UnsignedSpend
}

type UnsignedSpend struct {
	Coin     *types.Coin
	Solution []byte
	Message  string
}
