package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/chia-network/go-chia-libs/pkg/types"
)

type Client struct {
	fullNodeService *FullNodeService
	walletService   *WalletService
}

func NewClient(host string, port uint16) *Client {
	return &Client{
		fullNodeService: DefaultFullNodeService(host, port),
		walletService:   DefaultWalletService(host, port),
	}
}

func (cli *Client) GetSyncStatus() (bool, error) {
	resp, httpResp, err := cli.fullNodeService.GetBlockchainState()
	if err != nil {
		return false, err
	}

	if httpResp.StatusCode != 200 {
		return false, fmt.Errorf("failed to request,status code:%v", httpResp.StatusCode)
	}

	if resp == nil || resp.BlockchainState.ToPointer() == nil {
		return false, fmt.Errorf("cannot get response from node")
	}

	syncState := resp.BlockchainState.ToPointer().Sync
	return syncState.Synced, nil
}

func (cli *Client) GetBalance(puzzleHash types.Bytes32) (uint64, error) {
	resp, httpResp, err := cli.fullNodeService.GetCoinRecordsByPuzzleHash(&GetCoinRecordsByPuzzleHashOptions{
		PuzzleHash:        puzzleHash,
		IncludeSpentCoins: false,
	})

	if err != nil {
		return 0, err
	}

	if httpResp.StatusCode != 200 {
		return 0, fmt.Errorf("failed to request,status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return 0, fmt.Errorf("cannot get response from node")
	}

	total := uint64(0)
	_total := uint64(0) // test for overflow
	for _, records := range resp.CoinRecords {
		total += records.Coin.Amount
		if total < _total {
			return math.MaxUint64, nil
		}
		_total = total
	}

	return total, nil
}

func (cli *Client) SelectCoins(totalAmount uint64, puzzleHash types.Bytes32) ([]*types.Coin, error) {
	resp, httpResp, err := cli.fullNodeService.GetCoinRecordsByPuzzleHash(&GetCoinRecordsByPuzzleHashOptions{
		PuzzleHash:        puzzleHash,
		IncludeSpentCoins: false,
	})

	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to request,status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return nil, fmt.Errorf("cannot get response from node")
	}

	return selectCoins(totalAmount, resp.CoinRecords)
}

func (cli *Client) PushTX(SpendBundle types.SpendBundle) (string, error) {
	resp, httpResp, err := cli.fullNodeService.PushTX(&FullNodePushTXOptions{SpendBundle})
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != 200 {
		return "", fmt.Errorf("failed to request, status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return "", fmt.Errorf("cannot get response from node")
	}

	if !resp.Success {
		return "", fmt.Errorf("failed to push tx, status:%v", resp.Status)
	}

	return TxHash(SpendBundle)
}

func selectCoins(totalAmount uint64, coins []types.CoinRecord) ([]*types.Coin, error) {
	aimCoins := []*types.Coin{}
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
func TxHash(spendBundele types.SpendBundle) (string, error) {
	txHashList := make([]string, 0)
	txHashList = append(txHashList, formateUint64(uint64(len(spendBundele.CoinSpends))))
	for _, cs := range spendBundele.CoinSpends {
		txHashList = append(txHashList,
			hex.EncodeToString(types.Bytes32ToBytes(cs.Coin.ParentCoinInfo)))
		txHashList = append(txHashList,
			hex.EncodeToString(types.Bytes32ToBytes(cs.Coin.PuzzleHash)))
		txHashList = append(txHashList,
			toBigEndingHex(cs.Coin.Amount))
		txHashList = append(txHashList,
			hex.EncodeToString([]byte(cs.PuzzleReveal)))
		txHashList = append(txHashList,
			hex.EncodeToString([]byte(cs.Solution)))
	}
	txHashList = append(txHashList,
		hex.EncodeToString(types.Bytes96ToBytes(types.Bytes96(spendBundele.AggregatedSignature))))
	streamTX := strings.Join(txHashList, "")
	fmt.Println(PrettyStruct(txHashList))
	streamTX = strings.ReplaceAll(streamTX, "0x", "")
	m, err := hex.DecodeString(streamTX)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(m)
	return hex.EncodeToString(hash[:]), nil
}

func toBigEndingHex(num uint64) string {
	hexStr := fmt.Sprintf("0000000000000000%x", num)
	hexStr = hexStr[len(hexStr)-16:]
	return hexStr
}

func formateUint64(num uint64) string {
	hexStr := fmt.Sprintf("00000000%x", num)
	hexStr = hexStr[len(hexStr)-8:]
	return hexStr
}

func PrettyStruct(data interface{}) string {
	val, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(val)
}
