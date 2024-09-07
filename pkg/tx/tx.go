package tx

import (
	"crypto/sha256"
	"fmt"

	wlog "github.com/NpoolPlatform/go-service-framework/pkg/wlog"

	puzzle1 "github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	rpc "github.com/NpoolPlatform/chia-client/pkg/rpc"
	types "github.com/NpoolPlatform/chia-client/pkg/types"
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/shopspring/decimal"
)

type txHandler struct {
	payments                   []*types.Payment
	amount                     *decimal.Decimal
	fee                        *decimal.Decimal
	totalBalance               *decimal.Decimal
	newPuzzleHash              *string
	changePuzzleHash           *string
	coins                      []*types1.Coin
	primaryCoin                *types1.Coin
	paymentAddition            *string
	createAnnouncementAddition *string
	assertAnnouncementAddition *string
	announcementMessage        *string
	announcementID             *string
	puzzleReveal               *string
	spends                     []*types.CoinSpend
	client                     *rpc.MyWalletService
}

func (h *txHandler) checkBalance() error {
	totalAmount := h.amount.Add(*h.fee)
	if err := h.getSpendableBalance(); err != nil {
		return wlog.WrapError(err)
	}
	if totalAmount.Cmp(*h.totalBalance) > 0 {
		return wlog.Errorf("insufficient funds")
	}
	return nil
}

func (h *txHandler) selectCoins(amount decimal.Decimal) error {
	totalAmount := h.amount.Add(*h.fee)
	coins, err := h.client.SelectCoins(&rpc.WalletOption{
		WalletID: 1,
		Amount:   totalAmount.BigInt().Uint64(),
	})
	if err != nil {
		return wlog.WrapError(err)
	}
	if len(coins) == 0 {
		return wlog.Errorf("coin not available")
	}
	h.primaryCoin = coins[0]
	h.coins = coins
	return nil
}

func (h *txHandler) getSpendableBalance() error {
	wallet, err := h.client.GetWalletBalance(&rpc.GetWalletBalanceOptions{
		WalletID: 1,
	})
	if err != nil {
		return wlog.WrapError(err)
	}

	totalBalance, err := decimal.NewFromString(wallet.SpendableBalance.String())
	if err != nil {
		return wlog.WrapError(err)
	}
	h.totalBalance = &totalBalance
	return nil
}

// TODO
func (h *txHandler) getNewPuzzleHash() (string, error) {
	return "0x8acc5d7f1383b5ea598d2f662c716ae8749730b736b58b24d0c8bae1497bc165", nil
}

// TODO:FinalPublicKey
func (h *txHandler) generatePuzzleReveal() error {
	_bytes, err := types1.BytesFromHexString("0xaf0c070c1ce82596d6f7450a7234ab663976e561f7e0676d746f9e332a0abef767bf3f2bd60b42d5977c0cfff1c06556")
	if err != nil {
		return wlog.WrapError(err)
	}
	puzzleReveal := puzzle1.NewProgramString(_bytes)
	h.puzzleReveal = &puzzleReveal
	return nil
}

func (h *txHandler) getComplementRepresentation(amount string) (string, error) {
	number, err := decimal.NewFromString(amount)
	if err != nil {
		return "", err
	}
	if number.Sign() < 0 {
		return "", wlog.Errorf("invalid amount")
	}

	bigInt := number.BigInt()
	bytes := bigInt.Bytes()

	if len(bytes) > 0 && bytes[0] >= 0x80 {
		bytes = append([]byte{0x00}, bytes...)
	}

	hexRepresentation := ""
	for _, b := range bytes {
		hexRepresentation += fmt.Sprintf("%02X", b)
	}

	prefix := "ff"
	if len(bytes) > 1 {
		prefix = fmt.Sprintf("ff8%d", len(bytes))
	}

	return prefix + hexRepresentation, nil
}

func (h *txHandler) stdHash(myBytes []types1.Bytes32) string {
	newBytes := sha256.New()
	for _, b := range myBytes {
		newBytes.Write(b[:])
	}
	_bytes := fmt.Sprintf("%x", newBytes.Sum(nil))
	return _bytes
}

func (h *txHandler) generatePayments() error {
	spendAmount := decimal.NewFromInt(0)
	for _, coin := range h.coins {
		coinAmount := fmt.Sprintf("%d", coin.Amount)
		spendAmount = spendAmount.Add(decimal.RequireFromString(coinAmount))
	}

	totalAmount := h.amount.Add(*h.fee)
	change := spendAmount.Sub(totalAmount)
	if change.Cmp(decimal.NewFromInt(0)) < 0 {
		return wlog.Errorf("negative change %s", change)
	}

	_newPuzzleHash, err := types1.Bytes32FromHexString(*h.newPuzzleHash)
	if err != nil {
		return wlog.WrapError(err)
	}
	payments := []*types.Payment{{
		PuzzleHash: _newPuzzleHash,
		Amount:     h.amount.String(),
	}}

	if change.Cmp(decimal.NewFromInt(0)) > 0 {
		changePuzzleHash, err := h.getNewPuzzleHash()
		if err != nil {
			return wlog.WrapError(err)
		}
		_changePuzzleHash, err := types1.Bytes32FromHexString(changePuzzleHash)
		if err != nil {
			return wlog.WrapError(err)
		}
		payments = append(payments, &types.Payment{
			PuzzleHash: _changePuzzleHash,
			Amount:     change.String(),
		})
	}
	h.payments = payments
	return nil
}

func (h *txHandler) generatePaymentAddition() error {
	addition := ""
	for _, payment := range h.payments {
		presentation, err := h.getComplementRepresentation(payment.Amount)
		if err != nil {
			return wlog.WrapError(err)
		}
		addition += "80ffff33ffa0" + payment.PuzzleHash.String()[2:] + presentation
	}
	if h.fee.Sign() > 0 {
		presentation, err := h.getComplementRepresentation(h.fee.String())
		if err != nil {
			return wlog.WrapError(err)
		}
		addition += "80ffff34" + presentation
	}
	h.paymentAddition = &addition
	return nil
}

func (h *txHandler) generateAnnouncementMessage() error {
	messages := []types1.Bytes32{}
	for _, coin := range h.coins {
		messages = append(messages, coin.ID())
	}

	for _, payment := range h.payments {
		amount, err := decimal.NewFromString(payment.Amount)
		if err != nil {
			return wlog.WrapError(err)
		}
		c := types1.Coin{
			ParentCoinInfo: h.primaryCoin.ID(),
			PuzzleHash:     payment.PuzzleHash,
			Amount:         amount.BigInt().Uint64(),
		}
		messages = append(messages, c.ID())
	}
	message := h.stdHash(messages)
	h.announcementMessage = &message
	return nil
}

func (h *txHandler) generateCreateAnnouncementAddition() {
	announcement := "ff01ffff3cffa0" + *h.announcementMessage
	h.createAnnouncementAddition = &announcement
}

func (h *txHandler) getAnnouncementID() error {
	message, err := types1.Bytes32FromHexString(*h.announcementMessage)
	if err != nil {
		return wlog.WrapError(err)
	}
	announcementID := h.stdHash([]types1.Bytes32{h.primaryCoin.ID(), message})
	h.announcementID = &announcementID
	return nil
}

func (h *txHandler) generateAssertAnnouncementAddition() {
	announcement := "ff01ffff3cffa0" + *h.announcementID
	h.assertAnnouncementAddition = &announcement
}

func (h *txHandler) formalize() {
	h.spends = append(h.spends, &types.CoinSpend{
		Coin:         *h.primaryCoin,
		PuzzleReveal: *h.puzzleReveal,
		Solution:     "0xff80ff" + *h.createAnnouncementAddition + *h.paymentAddition + "8080ff8080",
	})
	if len(h.coins) <= 1 {
		return
	}

	for _, coin := range h.coins[1:] {
		h.spends = append(h.spends, &types.CoinSpend{
			Coin:         *coin,
			PuzzleReveal: *h.puzzleReveal,
			Solution:     "0xff80ff" + *h.assertAnnouncementAddition + "8080ff8080",
		})
	}
}

func GenerateUnsignedTransaction(amount string, newPuzzleHash string, fee string) ([]*types.CoinSpend, error) {
	_amount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	_fee, err := decimal.NewFromString(fee)
	if err != nil {
		return nil, wlog.WrapError(err)
	}

	client, err := rpc.GetWalletClient()
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	txHandler := &txHandler{
		amount:        &_amount,
		fee:           &_fee,
		newPuzzleHash: &newPuzzleHash,
		client:        client,
	}
	if err := txHandler.checkBalance(); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.selectCoins(_amount); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.generatePuzzleReveal(); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.generatePayments(); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.generatePaymentAddition(); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.generateAnnouncementMessage(); err != nil {
		return nil, wlog.WrapError(err)
	}
	if err := txHandler.getAnnouncementID(); err != nil {
		return nil, wlog.WrapError(err)
	}
	txHandler.generateCreateAnnouncementAddition()
	txHandler.generateAssertAnnouncementAddition()
	txHandler.formalize()
	return txHandler.spends, nil
}
