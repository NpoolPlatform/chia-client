package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	wlog "github.com/NpoolPlatform/go-service-framework/pkg/wlog"

	puzzlehash1 "github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	types "github.com/NpoolPlatform/chia-client/pkg/types"
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/shopspring/decimal"
)

type txHandler struct {
	payments                   []*types.Payment
	amount                     *decimal.Decimal
	changeAmount               *decimal.Decimal
	fee                        *decimal.Decimal
	toPuzzleHash               *string
	toPuzzleHashBytes          []byte
	fromPuzzleHash             *string
	fromPuzzleHashBytes        []byte
	changePuzzleHash           *string
	records                    []*types1.CoinRecord
	coins                      []*types1.Coin
	primaryCoin                *types1.Coin
	paymentAddition            *string
	createAnnouncementAddition *string
	assertAnnouncementAddition *string
	announcementMessage        *string
	announcementID             *string
	puzzleReveal               *string
	spends                     []*types.CoinSpend
	additionalData             *string
	publicKey                  *string
	tx                         *types.UnsignedTx
}

func (h *txHandler) selectCoins(totalSpend decimal.Decimal) error {
	coins := []*types1.Coin{}
	spendableBalance := decimal.NewFromInt(0)

	for _, record := range h.records {
		amount := fmt.Sprintf("%d", record.Coin.Amount)
		spendableBalance = spendableBalance.Add(decimal.RequireFromString(amount))
		coins = append(coins, &record.Coin)
		if spendableBalance.Cmp(totalSpend) >= 0 {
			break
		}
	}
	if len(coins) == 0 {
		return wlog.Errorf("invalid coin")
	}
	if spendableBalance.Cmp(totalSpend) < 0 {
		return wlog.Errorf("insufficient funds")
	}
	h.primaryCoin = coins[0]
	h.coins = coins
	return nil
}

func (h *txHandler) generatePuzzleReveal() error {
	_bytes, err := types1.BytesFromHexString(*h.publicKey)
	if err != nil {
		return wlog.WrapError(err)
	}

	puzzleReveal := "0x" + puzzlehash1.NewProgramString(_bytes)
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
	h.changeAmount = &change

	payments := []*types.Payment{{
		PuzzleHash: types1.Bytes32(h.toPuzzleHashBytes),
		Amount:     h.amount.String(),
	}}

	if change.Sign() > 0 {
		payments = append(payments, &types.Payment{
			PuzzleHash: h.primaryCoin.PuzzleHash,
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
	messages := []string{"0x" + h.conditionChangeTreeHash() + h.primaryCoin.ID().String()[2:] + *h.additionalData}
	h.spends = append(h.spends, &types.CoinSpend{
		Coin:         *h.primaryCoin,
		PuzzleReveal: *h.puzzleReveal,
		Solution:     "0xff80ff" + *h.createAnnouncementAddition + *h.paymentAddition + "8080ff8080",
	})
	h.tx = &types.UnsignedTx{
		CoinSpends: h.spends,
		Messages:   messages,
	}
	if len(h.coins) < 2 {
		return
	}

	for _, coin := range h.coins[1:] {
		messages = append(messages, "0x"+h.conditionAssertTreeHash()+coin.ID().String()[2:]+*h.additionalData)
		h.spends = append(h.spends, &types.CoinSpend{
			Coin:         *coin,
			PuzzleReveal: *h.puzzleReveal,
			Solution:     "0xff80ff" + *h.assertAnnouncementAddition + "8080ff8080",
		})
	}
	h.tx = &types.UnsignedTx{
		CoinSpends: h.spends,
		Messages:   messages,
	}
}

func GenerateUnsignedTransaction(from string, to string, amount string, fee string, records []*types1.CoinRecord, additionalData string, publicKey string) (*types.UnsignedTx, error) {
	_amount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	_fee, err := decimal.NewFromString(fee)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	_, fromPuzzleHashBytes, err := puzzlehash1.GetPuzzleHashFromAddress(from)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	_, toPuzzleHashBytes, err := puzzlehash1.GetPuzzleHashFromAddress(to)
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	fromPuzzleHash := fmt.Sprintf("0x%s", hex.EncodeToString(fromPuzzleHashBytes))
	toPuzzleHash := fmt.Sprintf("0x%s", hex.EncodeToString(toPuzzleHashBytes))

	txHandler := &txHandler{
		amount:              &_amount,
		fee:                 &_fee,
		fromPuzzleHash:      &fromPuzzleHash,
		changePuzzleHash:    &fromPuzzleHash,
		fromPuzzleHashBytes: fromPuzzleHashBytes,
		toPuzzleHash:        &toPuzzleHash,
		toPuzzleHashBytes:   toPuzzleHashBytes,
		records:             records,
		additionalData:      &additionalData,
		publicKey:           &publicKey,
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
	return txHandler.tx, nil
}
