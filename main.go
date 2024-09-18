package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/NpoolPlatform/chia-client/pkg/account"
	"github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

const (
	aggsig_data = "ccd5bb71183532bff220ba46c268991a3ff07eb358e8255a65c30a2dce0e5fbb"
)

type TxInfo struct {
	From                string
	To                  string
	Amount              uint64
	Change              uint64
	Fee                 uint64
	AggregatedSignature string
	Spends              []*Spend
}

type Spend struct {
	Coin         *types.Coin `json:"coin" streamable:""`
	PuzzleReveal string      `json:"puzzle_reveal" streamable:""`
	Solution     string      `json:"solution" streamable:""`
	Message      string      `json:"message" streamable:""`
}

func main() {
	TxDemo()
	// TestSign()
	// TestCreateSolution()
	// TestPuzzleReveal()
	// TestTreeHash()

}

func TestSign() {
	fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	// fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	fromAcc, _ := account.GenAccountBySKHex(fromSKHex)
	fmt.Println(fromAcc.GetPKHex())
	fmt.Println(fromAcc.GetAddress(false))
	fmt.Println(hex.EncodeToString(fromAcc.Sign([]byte("hello"))))
}

func TxDemo() {
	// fromAcc, err := account.GenAccount()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(fromAcc.GetSKHex())
	// fmt.Println(fromAcc.GetPKHex())
	// fmt.Println(fromAcc.GetAddress(true))

	// toSKHex:=34215bfd598d6b1db74b27048aecf338240fee5c47555fdb5e19a390c145d5d0
	// toPKHex:=a6bdb8fb599e40b619d329667012a629c487f0cdd60c7159b571c19b439f407c48ebc796ade9e5968bbd09a40721bf25
	// toAddress := "txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz"

	// toAmount := uint64(100000001)

	fromSKHex := "3fefe074898e3ac7c6c17a40ec390d7c4ade53fde6c39339a93d03012bd3b7f7"
	// fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt"
	fromAcc, err := account.GenAccountBySKHex(fromSKHex)
	if err != nil {
		fmt.Println(1, err)
		return
	}

	fmt.Println(fromAcc.GetAddress(false))

	txInfo := TxInfo{
		From:   "txch1y2vqher2radvvkspad9l46jrewv63tm3huv9ewl2d37594eg3lrqtrlkgt",
		To:     "txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz",
		Amount: 66,
		Fee:    100,
	}

	_, fromPH, err := puzzlehash.GetPuzzleHashFromAddress(txInfo.From)
	if err != nil {
		fmt.Println(2, err)
		return
	}

	_, toPH, err := puzzlehash.GetPuzzleHashFromAddress(txInfo.To)
	if err != nil {
		fmt.Println(3, err)
		return
	}

	cli, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig(), rpc.WithBaseURL(&url.URL{Scheme: "https", Host: "localhost"}))
	if err != nil {
		fmt.Println(4, err)
		return
	}

	coinsRsp, _, err := cli.FullNodeService.GetCoinRecordsByPuzzleHash(&rpc.GetCoinRecordsByPuzzleHashOptions{
		PuzzleHash:        *fromPH,
		IncludeSpentCoins: false,
	})
	if err != nil {
		fmt.Println(5, err)
		return
	}

	selectedCoins, err := selectCoins(txInfo.Amount, txInfo.Fee, coinsRsp.CoinRecords)
	if err != nil {
		fmt.Println(5, err)
		return
	}

	paymentCoins, change, err := calPaymentCoins(txInfo.Amount, txInfo.Fee, txInfo.From, txInfo.To, selectedCoins)
	if err != nil {
		fmt.Println(6, err)
		return
	}
	txInfo.Change = change

	createAnnounceMSG := genCreateAnoucementMessage(selectedCoins, paymentCoins)

	createSolution, err := genCreateSolution(createAnnounceMSG, paymentCoins, txInfo.Fee)
	if err != nil {
		fmt.Println(7, err)
		return
	}

	assertSolution, err := genAssertSolution(createAnnounceMSG, selectedCoins[0])
	if err != nil {
		fmt.Println(8, err)
		return
	}

	pkBytes, err := fromAcc.GetPKBytes()
	if err != nil {
		fmt.Println(9, err)
		return
	}

	spends := []*Spend{
		{
			Coin:         selectedCoins[0],
			Solution:     createSolution,
			PuzzleReveal: puzzlehash.NewProgramString(pkBytes),
			Message: genUnsignedCreateMessage(
				createAnnounceMSG,
				types.Bytes32ToBytes(*toPH),
				types.Bytes32ToBytes(*fromPH),
				txInfo.Amount,
				txInfo.Change,
				txInfo.Fee,
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
			&Spend{
				Coin:         coin,
				Solution:     assertSolution,
				PuzzleReveal: puzzlehash.NewProgramString(pkBytes),
				Message:      assertMsg,
			})
	}

	signs := [][]byte{}
	for _, spend := range spends {
		msg, err := hex.DecodeString(spend.Message)
		if err != nil {
			fmt.Println(11, err)
			return
		}
		signs = append(signs, fromAcc.Sign(msg))
	}

	aggregateSign, err := account.AggregateSigns(signs)
	if err != nil {
		fmt.Println(12, err)
		return
	}

	fmt.Println(hex.EncodeToString(aggregateSign))
	fmt.Println(PrettyStruct(spends))
}

func TestPuzzleReveal() {
	skHexStr := "1c6198abdad4569b09554e48abc7f78d2c2833ed8235b862171a0ecf9db62d51"
	acc, err := account.GenAccountBySKHex(skHexStr)
	fmt.Println(err)
	fmt.Println(acc.GetPKHex())

	pkHex, err := hex.DecodeString("b8d50671a208e33f1fd8f85b664f5776a106a3f0c615da5068ca0fc153be606622bc4ac928ef3cf8283241ef4f44a866")

	fmt.Println(puzzlehash.NewProgramString(pkHex))
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

func genAssertSolution(createAnnounceMSG []byte, firstCoin *types.Coin) (string, error) {
	msgSum := sha256.New()
	_, err := msgSum.Write(types.Bytes32ToBytes(firstCoin.ID()))
	if err != nil {
		return "", err
	}

	_, err = msgSum.Write(createAnnounceMSG)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"ff80ffff01ffff3dffa0%v8080ff8080",
		hex.EncodeToString(msgSum.Sum(nil)),
	), nil
}

func genCreateSolution(createAnnounceMSG []byte, paymentCoins []*types.Coin, fee uint64) (string, error) {
	if len(paymentCoins) != 2 {
		return "", fmt.Errorf("invalid payment coins")
	}

	return fmt.Sprintf(
		"ff80ffff01ffff3cffa0%v80ffff33ffa0%vff%v80ffff33ffa0%vff%v80ffff34ff%v8080ff8080",
		hex.EncodeToString(createAnnounceMSG),
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[0].PuzzleHash)),
		hex.EncodeToString(encodeU64ToCLVMBytes(paymentCoins[0].Amount)),
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[1].PuzzleHash)),
		hex.EncodeToString(encodeU64ToCLVMBytes(paymentCoins[1].Amount)),
		hex.EncodeToString(encodeU64ToCLVMBytes(fee)),
	), nil
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
