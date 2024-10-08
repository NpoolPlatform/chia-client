package transaction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/NpoolPlatform/chia-client/pkg/client"
	"github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

func GenSignedSpendBundle(unsignedTx *UnsignedTx, fromSKHex string) (*types.SpendBundle, error) {
	fromAcc, err := account.GenAccountBySKHex(fromSKHex)
	if err != nil {
		return nil, fmt.Errorf("invalid sk,err: %v", err)
	}

	pkBytes, err := fromAcc.GetPKBytes()
	if err != nil {
		return nil, fmt.Errorf("cannot get pk from sk,err: %v", err)
	}

	signedSpends := []types.CoinSpend{}

	signs := [][]byte{}
	for _, spend := range unsignedTx.Spends {
		msg, err := hex.DecodeString(spend.Message)
		if err != nil {
			return nil, fmt.Errorf("wrong message,err: %v", err)
		}

		signedSpends = append(signedSpends, types.CoinSpend{
			Coin:         *spend.Coin,
			PuzzleReveal: puzzlehash.NewProgramBytes(pkBytes),
			Solution:     spend.Solution,
		})
		signs = append(signs, fromAcc.Sign(msg))
	}

	aggregateSign, err := account.AggregateSigns(signs)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate signatures,err: %v", err)
	}

	aggSign, err := types.BytesToBytes96(aggregateSign)
	if err != nil {
		return nil, fmt.Errorf("wrong aggregated signature,err: %v", err)
	}

	spendBundle := types.SpendBundle{
		AggregatedSignature: types.G2Element(aggSign),
		CoinSpends:          signedSpends,
	}
	return &spendBundle, nil
}

func GenUnsignedTx(ctx context.Context, cli *client.Client, from, to string, amount, fee uint64) (*UnsignedTx, error) {
	unsignedTx := &UnsignedTx{
		From: from,
	}

	_, fromPH, err := puzzlehash.GetPuzzleHashFromAddress(from)
	if err != nil {
		return nil, fmt.Errorf("invalid format for from address,err: %v", err)
	}

	_, toPH, err := puzzlehash.GetPuzzleHashFromAddress(to)
	if err != nil {
		return nil, fmt.Errorf("invalid format for to address,err: %v", err)
	}

	selectedCoins, err := cli.SelectCoins(ctx, amount+fee, *fromPH)
	if err != nil {
		return nil, fmt.Errorf("failed to select coins,err: %v", err)
	}

	for _, coin := range selectedCoins {
		unsignedTx.SpentCoinIDs = append(unsignedTx.SpentCoinIDs, coin.ID().String())
	}

	paymentCoins, change, err := calPaymentCoins(amount, fee, from, to, selectedCoins)
	if err != nil {
		return nil, fmt.Errorf("failed to cal payment coins,err: %v", err)
	}

	createAnnounceMSG := genCreateAnoucementMessage(selectedCoins, paymentCoins)

	createSolution, err := genCreateSolution(createAnnounceMSG, paymentCoins, fee)
	if err != nil {
		return nil, fmt.Errorf("failed to generate create solution,err: %v", err)
	}

	assertSolution, err := genAssertSolution(createAnnounceMSG, selectedCoins[0])
	if err != nil {
		return nil, fmt.Errorf("failed to generate assert solution,err: %v", err)
	}

	aggsigData, err := cli.GetAggsigAddtionalData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggsig data,err: %v", err)
	}

	spends := []*UnsignedSpend{
		{
			Coin:     selectedCoins[0],
			Solution: createSolution,
			Message: genUnsignedCreateMessage(
				createAnnounceMSG,
				types.Bytes32ToBytes(*toPH),
				types.Bytes32ToBytes(*fromPH),
				amount,
				change,
				fee,
				selectedCoins[0],
				*aggsigData,
			),
		},
	}

	for _, coin := range selectedCoins[1:] {
		assertMsg, err := genUnsignedAssertMessage(createAnnounceMSG, selectedCoins[0], coin, *aggsigData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate unsigned assert message,err: %v", err)
		}
		spends = append(spends,
			&UnsignedSpend{
				Coin:     coin,
				Solution: assertSolution,
				Message:  assertMsg,
			})
	}
	unsignedTx.Spends = spends

	return unsignedTx, nil
}

func genUnsignedCreateMessage(announceMsg, toPH, fromPH []byte, amount, change, fee uint64, firstCoin *types.Coin, aggsigData types.Bytes32) string {
	cCTreeHash := conditionCreateTreeHash(announceMsg, toPH, fromPH, amount, change, fee)
	unsignMsg := cCTreeHash + hex.EncodeToString(types.Bytes32ToBytes(firstCoin.ID())) + hex.EncodeToString(types.Bytes32ToBytes(aggsigData))
	return unsignMsg
}

func genUnsignedAssertMessage(announceMsg []byte, firstCoin, coin *types.Coin, aggsigData types.Bytes32) (string, error) {
	s := sha256.New()
	_, err := s.Write(append(types.Bytes32ToBytes(firstCoin.ID()), announceMsg...))
	if err != nil {
		return "", err
	}
	announcementID := s.Sum(nil)
	assertTreeHash := conditionAssertTreeHash(announcementID)

	unsignMsg := assertTreeHash + hex.EncodeToString(types.Bytes32ToBytes(coin.ID())) + hex.EncodeToString(types.Bytes32ToBytes(aggsigData))
	return unsignMsg, nil
}

func genAssertSolution(createAnnounceMSG []byte, firstCoin *types.Coin) ([]byte, error) {
	msgSum := sha256.New()
	_, err := msgSum.Write(types.Bytes32ToBytes(firstCoin.ID()))
	if err != nil {
		return nil, err
	}

	_, err = msgSum.Write(createAnnounceMSG)
	if err != nil {
		return nil, err
	}

	solutionStr := fmt.Sprintf(
		"ff80ffff01ffff3dffa0%v8080ff8080",
		hex.EncodeToString(msgSum.Sum(nil)),
	)

	return hex.DecodeString(solutionStr)
}

func genCreateSolution(createAnnounceMSG []byte, paymentCoins []*types.Coin, fee uint64) ([]byte, error) {
	if len(paymentCoins) != 2 {
		return nil, fmt.Errorf("invalid payment coins")
	}

	solutionStr := fmt.Sprintf(
		"ff80ffff01ffff3cffa0%v80ffff33ffa0%vff%v80ffff33ffa0%vff%v80ffff34ff%v8080ff8080",
		hex.EncodeToString(createAnnounceMSG),
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[0].PuzzleHash)),
		hex.EncodeToString(encodeU64ToCLVMBytes(paymentCoins[0].Amount)),
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[1].PuzzleHash)),
		hex.EncodeToString(encodeU64ToCLVMBytes(paymentCoins[1].Amount)),
		hex.EncodeToString(encodeU64ToCLVMBytes(fee)),
	)

	return hex.DecodeString(solutionStr)
}

func genCreateAnoucementMessage(selectedCoins, paymentCoins []*types.Coin) []byte {
	coinIDsBytes := sha256.New()
	for _, coin := range selectedCoins {
		coinIDsBytes.Write(types.Bytes32ToBytes(coin.ID()))
	}

	for _, coin := range paymentCoins {
		coinIDsBytes.Write(types.Bytes32ToBytes(coin.ID()))
	}

	return coinIDsBytes.Sum(nil)
}

func calPaymentCoins(amount, fee uint64, from, to string, selectedCoins []*types.Coin) ([]*types.Coin, uint64, error) {
	totalAmount := uint64(0)
	paymentAccount := amount + fee

	for _, coin := range selectedCoins {
		totalAmount += coin.Amount
	}

	if totalAmount < paymentAccount {
		return nil, 0, fmt.Errorf("amount is insufficient")
	}

	_, fromPHBytes, err := puzzlehash.GetPuzzleHashFromAddress(from)
	if err != nil {
		return nil, 0, err
	}

	_, toPHBytes, err := puzzlehash.GetPuzzleHashFromAddress(to)
	if err != nil {
		return nil, 0, err
	}

	changeAmount := totalAmount - paymentAccount

	// construct payment coin
	paymentCoin := &types.Coin{
		ParentCoinInfo: selectedCoins[0].ID(),
		PuzzleHash:     *toPHBytes,
		Amount:         amount,
	}

	// construct change coin
	changeCoin := &types.Coin{
		ParentCoinInfo: selectedCoins[0].ID(),
		PuzzleHash:     *fromPHBytes,
		Amount:         changeAmount,
	}

	return []*types.Coin{paymentCoin, changeCoin}, changeAmount, nil
}

func encodeU64ToCLVMBytes(amount uint64) []byte {
	hexAmount := fmt.Sprintf("%x", amount)
	if len(hexAmount)%2 != 0 {
		hexAmount = fmt.Sprintf("0%v", hexAmount)
	}
	if amount >= 0x80 {
		if hexAmount[:2] >= "80" {
			hexAmount = fmt.Sprintf("00%v", hexAmount)
		}
		hexAmount = fmt.Sprintf("8%x%v", len(hexAmount)/2, hexAmount)
	}
	hexAmountBytes, _ := hex.DecodeString(hexAmount)
	return hexAmountBytes
}

func encodeU64ToBytes(amount uint64) []byte {
	hexAmount := fmt.Sprintf("%x", amount)
	if len(hexAmount)%2 != 0 {
		hexAmount = fmt.Sprintf("0%v", hexAmount)
	}
	if hexAmount[:2] >= "80" {
		hexAmount = fmt.Sprintf("00%v", hexAmount)
	}
	hexAmountBytes, _ := hex.DecodeString(hexAmount)
	return hexAmountBytes
}

type treeNode struct {
	left  *treeNode
	right *treeNode
	val   []byte
}

func conditionCreateTreeHash(createAnnounceMSG, toPH, fromPH []byte, amount, change, fee uint64) string {
	tree := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{60}},
				right: &treeNode{
					left:  &treeNode{val: createAnnounceMSG},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{
				left: &treeNode{
					left: &treeNode{val: []byte{51}},
					right: &treeNode{
						left: &treeNode{val: toPH},
						right: &treeNode{
							left:  &treeNode{val: encodeU64ToBytes(amount)},
							right: &treeNode{val: []byte{}},
						},
					},
				},
				right: &treeNode{
					left: &treeNode{
						left: &treeNode{val: []byte{51}},
						right: &treeNode{
							left: &treeNode{val: fromPH},
							right: &treeNode{
								left:  &treeNode{val: encodeU64ToBytes(change)},
								right: &treeNode{val: []byte{}},
							},
						},
					},
					right: &treeNode{
						left: &treeNode{
							left: &treeNode{val: []byte{52}},
							right: &treeNode{
								left:  &treeNode{val: encodeU64ToBytes(fee)},
								right: &treeNode{val: []byte{}},
							},
						},
						right: &treeNode{val: []byte{}},
					},
				},
			},
		},
	}
	treeH := sha256tree(&tree)
	return hex.EncodeToString(treeH[:])
}

func conditionAssertTreeHash(announcementID []byte) string {
	tree1 := treeNode{
		left: &treeNode{val: []byte{1}},
		right: &treeNode{
			left: &treeNode{
				left: &treeNode{val: []byte{61}},
				right: &treeNode{
					left:  &treeNode{val: announcementID},
					right: &treeNode{val: []byte{}},
				},
			},
			right: &treeNode{val: []byte{}},
		},
	}

	treeH1 := sha256tree(&tree1)

	fmt.Println(hex.EncodeToString(treeH1[:]))
	return hex.EncodeToString(treeH1[:])
}

func sha256tree(v *treeNode) [32]byte {
	sBytes := []byte{}
	if v.left != nil {
		left := sha256tree(v.left)
		right := sha256tree(v.right)

		sBytes = append(sBytes, byte(2))
		sBytes = append(sBytes, left[:]...)
		sBytes = append(sBytes, right[:]...)
	} else {
		sBytes = append(sBytes, byte(1))
		sBytes = append(sBytes, v.val...)
	}

	return sha256.Sum256(sBytes)
}
