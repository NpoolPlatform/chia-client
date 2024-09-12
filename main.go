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
	*types.Coin
	Puzzle   string
	Solution string
	Message  string
}

func main() {
	TxDemo()
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
	fromPKHex := "b5cdc71cbceee853fdc397a209640097852496d2611c252c41477dc68ea54f2b507b9a34cc909f77a70ea06824774a3d"
	// fromAddress := "txch1vj27w3fngwqz6kwg5rmug9s6m5v8zc3nhkgwhg8pku88kav0pclsg4x8dj"
	fromAcc, err := account.GenAccountBySKHex(fromSKHex)
	if err != nil {
		fmt.Println(err)
		return
	}

	txInfo := TxInfo{
		From:   "txch1vj27w3fngwqz6kwg5rmug9s6m5v8zc3nhkgwhg8pku88kav0pclsg4x8dj",
		To:     "txch1pccwlj52r39yul8hp5mm3q96462up8k3xk83muwjyjhvy2vxnqwsnt40tz",
		Amount: 100000,
		Fee:    1,
	}

	_, fromPH, err := puzzlehash.GetPuzzleHashFromAddress(txInfo.From)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, toPH, err := puzzlehash.GetPuzzleHashFromAddress(txInfo.To)
	if err != nil {
		fmt.Println(err)
		return
	}

	cli, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig(), rpc.WithBaseURL(&url.URL{Scheme: "https", Host: "localhost"}))
	if err != nil {
		fmt.Println(err)
		return
	}

	coinsRsp, _, err := cli.FullNodeService.GetCoinRecordsByPuzzleHash(&rpc.GetCoinRecordsByPuzzleHashOptions{
		PuzzleHash:        *fromPH,
		IncludeSpentCoins: false,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	selectedCoins, err := selectCoins(txInfo.Amount, txInfo.Fee, coinsRsp.CoinRecords)
	if err != nil {
		fmt.Println(err)
		return
	}

	paymentCoins, change, err := calPaymentCoins(txInfo.Amount, txInfo.Fee, txInfo.From, txInfo.To, selectedCoins)
	if err != nil {
		fmt.Println(err)
		return
	}
	txInfo.Change = change

	createAnnounceMSG := genCreateAnoucementMessage(selectedCoins, paymentCoins)

	createSolution, err := genCreateSolution(createAnnounceMSG, paymentCoins, txInfo.Fee)
	if err != nil {
		fmt.Println(err)
		return
	}

	assertSolution, err := genAssertSolution(createAnnounceMSG, selectedCoins[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	spends := []*Spend{
		{
			Coin:     selectedCoins[0],
			Solution: createSolution,
			Puzzle:   puzzlehash.NewProgramString([]byte(fromPKHex)),
			Message: genUnsignedCreateMessage(
				createAnnounceMSG,
				types.Bytes32ToBytes(*fromPH),
				types.Bytes32ToBytes(*toPH),
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
			fmt.Println(err)
			return
		}
		spends = append(spends,
			&Spend{
				Coin:     coin,
				Solution: assertSolution,
				Puzzle:   puzzlehash.NewProgramString([]byte(fromPKHex)),
				Message:  assertMsg,
			})
	}

	fmt.Println(PrettyStruct(spends))
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

func genUnsignedCreateMessage(announceMsg string, fromPH, toPH []byte, amount, change, fee uint64, firstCoin *types.Coin) string {
	cCTreeHash := conditionCreateTreeHash([]byte(announceMsg), fromPH, toPH, amount, change, fee)
	unsignMsg := cCTreeHash + firstCoin.ID().String() + aggsig_data
	return unsignMsg
}

func genUnsignedAssertMessage(announceMsg string, firstCoin, coin *types.Coin) (string, error) {
	s := sha256.New()
	_, err := s.Write(append(types.Bytes32ToBytes(firstCoin.ID()), []byte(announceMsg)...))
	if err != nil {
		return "", err
	}
	announcementID := s.Sum(nil)
	assertTreeHash := conditionAssertTreeHash(announcementID)

	unsignMsg := assertTreeHash + coin.ID().String() + aggsig_data
	return unsignMsg, nil
}

func genAssertSolution(createAnnounceMSG string, firstCoin *types.Coin) (string, error) {
	msgSum := sha256.New()
	_, err := msgSum.Write(types.Bytes32ToBytes(firstCoin.ID()))
	if err != nil {
		return "", err
	}

	msgBytes, err := hex.DecodeString(createAnnounceMSG)
	if err != nil {
		return "", err
	}

	_, err = msgSum.Write(msgBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"ff80ffff01ffff3dffa0%v8080ff8080",
		hex.EncodeToString(msgSum.Sum(nil)),
	), nil
}

func genCreateSolution(createAnnounceMSG string, paymentCoins []*types.Coin, fee uint64) (string, error) {
	if len(paymentCoins) != 2 {
		return "", fmt.Errorf("invalid payment coins")
	}

	return fmt.Sprintf(
		"ff80ffff01ffff3cffa0%v80ffff33ffa0%vff%v80ffff33ffa0%vff%v80ffff34ff%v8080ff8080",
		createAnnounceMSG,
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[0].PuzzleHash)),
		encodeU64ToCLVMBytes(paymentCoins[0].Amount),
		hex.EncodeToString(types.Bytes32ToBytes(paymentCoins[1].PuzzleHash)),
		encodeU64ToCLVMBytes(paymentCoins[1].Amount),
		encodeU64ToCLVMBytes(fee),
	), nil
}

func genCreateAnoucementMessage(selectedCoins, paymentCoins []*types.Coin) string {
	coinIDsBytes := sha256.New()
	for _, coin := range selectedCoins {
		coinIDsBytes.Write(types.Bytes32ToBytes(coin.ID()))
	}

	for _, coin := range paymentCoins {
		coinIDsBytes.Write(types.Bytes32ToBytes(coin.ID()))
	}

	return hex.EncodeToString(coinIDsBytes.Sum(nil))
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
	// construct change coin
	changeCoin := &types.Coin{
		ParentCoinInfo: selectedCoins[0].ID(),
		PuzzleHash:     *fromPHBytes,
		Amount:         changeAmount,
	}

	// construct payment coin
	paymentCoin := &types.Coin{
		ParentCoinInfo: selectedCoins[0].ID(),
		PuzzleHash:     *toPHBytes,
		Amount:         amount,
	}

	return []*types.Coin{changeCoin, paymentCoin}, changeAmount, nil
}

func encodeU64ToCLVMBytes(amount uint64) string {
	hexAmount := fmt.Sprintf("%x", amount)
	if len(hexAmount)%2 != 0 {
		hexAmount = fmt.Sprintf("0%v", hexAmount)
	}

	if amount >= 0x80 {
		hexAmount = fmt.Sprintf("8%x%v", len(hexAmount)/2, hexAmount)
	}

	return hexAmount
}

type treeNode struct {
	left  *treeNode
	right *treeNode
	val   []byte
}

func conditionCreateTreeHash(createAnnounceMSG, fromPH, toPH []byte, amount, change, fee uint64) string {
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
							left:  &treeNode{val: []byte(encodeU64ToCLVMBytes(amount))},
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
								left:  &treeNode{val: []byte(encodeU64ToCLVMBytes(change))},
								right: &treeNode{val: []byte{}},
							},
						},
					},
					right: &treeNode{
						left: &treeNode{
							left: &treeNode{val: []byte{52}},
							right: &treeNode{
								left:  &treeNode{val: []byte(encodeU64ToCLVMBytes(fee))},
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
		sBytes = append(sBytes, []byte(v.val)...)
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
