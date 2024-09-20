package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/NpoolPlatform/chia-client/pkg/client"
	"github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	"github.com/chia-network/go-chia-libs/pkg/streamable"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

const (
	aggsig_data = "ccd5bb71183532bff220ba46c268991a3ff07eb358e8255a65c30a2dce0e5fbb"
)

type UnsignedTx struct {
	From   string
	Spends []*UnsignedSpend
}

type UnsignedSpend struct {
	Coin     *types.Coin
	Solution []byte
	Message  string
}

func main() {
	TxDemo()
	// TestCreateSolution()
	// TestTreeHash()
	// TestClient()
}

func TestClient() {
	cli := client.NewClient("172.16.31.202", 18444)
	fmt.Println(cli.GetSyncStatus())

	_, fromPH, err := puzzlehash.GetPuzzleHashFromAddress("txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz")
	if err != nil {
		fmt.Println(2, err)
		return
	}

	fmt.Println(cli.GetBalance(*fromPH))

	coinInfos, err := cli.SelectCoins(106732, *fromPH)
	fmt.Println(err)
	fmt.Println(PrettyStruct(coinInfos))
}

func TxDemo() {
	// ----------------------------Check Node Heath-----------------------------
	cli := client.NewClient("172.16.31.202", 18444)
	synced, err := cli.GetSyncStatus()
	if err != nil {
		fmt.Println(1, err)
		return
	}
	if !synced {
		fmt.Println("node have not synced")
		return
	}

	// ----------------------------Prepare UnsignedTX-----------------------------
	From := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	To := "txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz"
	Amount := uint64(66)
	Fee := uint64(100)

	unsignedTx := UnsignedTx{
		From: From,
	}

	_, fromPH, err := puzzlehash.GetPuzzleHashFromAddress(From)
	if err != nil {
		fmt.Println(2, err)
		return
	}

	_, toPH, err := puzzlehash.GetPuzzleHashFromAddress(To)
	if err != nil {
		fmt.Println(3, err)
		return
	}

	selectedCoins, err := cli.SelectCoins(Amount+Fee, *fromPH)
	if err != nil {
		fmt.Println(5, err)
		return
	}

	paymentCoins, Change, err := calPaymentCoins(Amount, Fee, From, To, selectedCoins)
	if err != nil {
		fmt.Println(6, err)
		return
	}

	createAnnounceMSG := genCreateAnoucementMessage(selectedCoins, paymentCoins)

	createSolution, err := genCreateSolution(createAnnounceMSG, paymentCoins, Fee)
	if err != nil {
		fmt.Println(7, err)
		return
	}

	assertSolution, err := genAssertSolution(createAnnounceMSG, selectedCoins[0])
	if err != nil {
		fmt.Println(8, err)
		return
	}

	spends := []*UnsignedSpend{
		{
			Coin:     selectedCoins[0],
			Solution: createSolution,
			Message: genUnsignedCreateMessage(
				createAnnounceMSG,
				types.Bytes32ToBytes(*toPH),
				types.Bytes32ToBytes(*fromPH),
				Amount,
				Change,
				Fee,
				selectedCoins[0],
			),
		},
	}

	for _, coin := range selectedCoins[1:] {
		assertMsg, err := genUnsignedAssertMessage(createAnnounceMSG, selectedCoins[0], coin)
		if err != nil {
			fmt.Println(10, err)
			return
		}

		spends = append(spends,
			&UnsignedSpend{
				Coin:     coin,
				Solution: assertSolution,
				Message:  assertMsg,
			})
	}

	unsignedTx.Spends = spends

	// ----------------------------SignTx-----------------------------
	// fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	// fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	fromAcc, err := account.GenAccountBySKHex(fromSKHex)
	if err != nil {
		fmt.Println(1, err)
		return
	}

	pkBytes, err := fromAcc.GetPKBytes()
	if err != nil {
		fmt.Println(9, err)
		return
	}

	signedSpends := []types.CoinSpend{}
	_signedSpends := []CoinSpend{}

	signs := [][]byte{}
	for _, spend := range unsignedTx.Spends {
		msg, err := hex.DecodeString(spend.Message)
		if err != nil {
			fmt.Println(11, err)
			return
		}

		signedSpends = append(signedSpends, types.CoinSpend{
			Coin:         *spend.Coin,
			PuzzleReveal: puzzlehash.NewProgramBytes(pkBytes),
			Solution:     spend.Solution,
		})

		_signedSpends = append(_signedSpends, CoinSpend{
			Coin:         *spend.Coin,
			PuzzleReveal: puzzlehash.NewProgramBytes(pkBytes),
			Solution:     spend.Solution,
		})
		signs = append(signs, fromAcc.Sign(msg))
	}

	aggregateSign, err := account.AggregateSigns(signs)
	if err != nil {
		fmt.Println(12, err)
		return
	}

	aggSign, err := types.BytesToBytes96(aggregateSign)
	if err != nil {
		fmt.Println(13, err)
		return
	}

	spendBundle := types.SpendBundle{
		AggregatedSignature: types.G2Element(aggSign),
		CoinSpends:          signedSpends,
	}

	_spendBundle := SpendBundle{
		AggregatedSignature: aggSign[:],
		CoinSpends:          _signedSpends,
	}

	sssss, err := streamable.Marshal(_spendBundle)
	if err != nil {
		fmt.Println(14, err)
		return
	}
	_sssss := sha256.Sum256(sssss)
	fmt.Println(hex.EncodeToString(_sssss[:]))
	// ----------------------------BroadcostTX-----------------------------
	fmt.Println(PrettyStruct(spendBundle))
	fmt.Println(cli.PushTX(spendBundle))
}

type SpendBundle struct {
	CoinSpends          []CoinSpend `json:"coin_spends" streamable:""`
	AggregatedSignature []byte      `json:"aggregated_signature" streamable:""`
}
type CoinSpend struct {
	Coin         types.Coin `json:"coin" streamable:""`
	PuzzleReveal []byte     `json:"puzzle_reveal" streamable:"SerializedProgram"`
	Solution     []byte     `json:"solution" streamable:"SerializedProgram"`
}

func TestCreateSolution() {
	var toB32 = func(a string) types.Bytes32 {
		b32, _ := types.Bytes32FromHexString(a)
		return b32
	}

	selectedCoins := []*types.Coin{
		{
			Amount:         9999,
			ParentCoinInfo: toB32("0x172bd20195e397e8c61ba46b11dc9e1c8882d4edc5e122e5d48b14079137385e"),
			PuzzleHash:     toB32("0xd16567930e5297d8d164b49ac3f9c20db715b98e53de7aafea07ba8fef306ba8"),
		},
	}

	fmt.Println(puzzlehash.GetAddressFromPuzzleHash(types.Bytes32ToBytes(toB32("0xa0daf9ee8ce0b2012557d99d42d9c36731d27b9cbc01848dffac39f001473eac")), "txch"))
	paymentCoins, _, err := calPaymentCoins(666, 1, "txch15rd0nm5vuzeqzf2hmxw59kwrvucay7uuhsqcfr0l4sulqq2886kqa5wtu0", "txch1cy7ru3sqvda8eft3394kuj6eaddhjuylnw3a99469878zm8q6azqxl8lv2", selectedCoins)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(PrettyStruct(paymentCoins))

	// paymentCoins := []*types.Coin{
	// 	{
	// 		Amount:         666,
	// 		ParentCoinInfo: toB32("0xee25937bd799c2441dedaea4cd37be49a160a4d6f0079c2efcbb1b417f7e0828"),
	// 		PuzzleHash:     toB32("0x6495e7453343802d59c8a0f7c4161add18716233bd90eba0e1b70e7b758f0e3f"),
	// 	},
	// 	{
	// 		Amount:         9332,
	// 		ParentCoinInfo: toB32("0xee25937bd799c2441dedaea4cd37be49a160a4d6f0079c2efcbb1b417f7e0828"),
	// 		PuzzleHash:     toB32("0xa0daf9ee8ce0b2012557d99d42d9c36731d27b9cbc01848dffac39f001473eac"),
	// 	},
	// }

	createAnnounceMSG := genCreateAnoucementMessage(selectedCoins, paymentCoins)

	createSolution, err := genCreateSolution(createAnnounceMSG, paymentCoins, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(createSolution)

	fmt.Println(genUnsignedCreateMessage(
		createAnnounceMSG,
		types.Bytes32ToBytes(paymentCoins[0].PuzzleHash),
		types.Bytes32ToBytes(paymentCoins[1].PuzzleHash),
		666,
		9332,
		1,
		selectedCoins[0],
	))
}

func TestTreeHash() {
	var toBytes = func(a string) []byte {
		b, _ := hex.DecodeString(a)
		return b
	}
	fmt.Println(conditionCreateTreeHash(
		toBytes("9c60986453d7a71d14d750026e742650d5f7f8a63ceccf08a1d12981caeb48bd"),
		toBytes("0e30efca8a1c4a4e7cf70d37b880baae95c09ed1358f1df1d224aec22986981d"),
		toBytes("6495e7453343802d59c8a0f7c4161add18716233bd90eba0e1b70e7b758f0e3f"),
		66,
		599,
		1,
	))
}

func selectCoins(amount, fee uint64, coins []types.CoinRecord) ([]*types.Coin, error) {
	aimCoins := []*types.Coin{}
	totalAmount := amount + fee
	sumAmount := uint64(0)

	for _, coin := range coins {
		sumAmount += coin.Coin.Amount
		aimCoins = append(aimCoins, &coin.Coin)
		if sumAmount >= totalAmount {
			break
		}
	}

	if sumAmount < totalAmount {
		return nil, fmt.Errorf("amount is insufficient")
	}

	return aimCoins, nil
}

func genUnsignedCreateMessage(announceMsg, toPH, fromPH []byte, amount, change, fee uint64, firstCoin *types.Coin) string {
	cCTreeHash := conditionCreateTreeHash(announceMsg, toPH, fromPH, amount, change, fee)
	unsignMsg := cCTreeHash + hex.EncodeToString(types.Bytes32ToBytes(firstCoin.ID())) + aggsig_data
	return unsignMsg
}

func genUnsignedAssertMessage(announceMsg []byte, firstCoin, coin *types.Coin) (string, error) {
	s := sha256.New()
	_, err := s.Write(append(types.Bytes32ToBytes(firstCoin.ID()), announceMsg...))
	if err != nil {
		return "", err
	}
	announcementID := s.Sum(nil)
	assertTreeHash := conditionAssertTreeHash(announcementID)

	unsignMsg := assertTreeHash + hex.EncodeToString(types.Bytes32ToBytes(coin.ID())) + aggsig_data
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

func PrettyStruct(data interface{}) string {
	val, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(val)
}
