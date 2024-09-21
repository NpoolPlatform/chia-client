package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/NpoolPlatform/chia-client/pkg/puzzlehash"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

type Client struct {
	fullNodeService *FullNodeService
	walletService   *WalletService
}

func NewClient(endpoint string) *Client {
	return &Client{
		fullNodeService: DefaultFullNodeService(endpoint),
		walletService:   DefaultWalletService(endpoint),
	}
}

func (cli *Client) GetSyncStatus(ctx context.Context) (bool, error) {
	resp, httpResp, err := cli.fullNodeService.GetBlockchainState(ctx)
	if err != nil {
		return false, err
	}

	if httpResp.StatusCode != 200 {
		return false, fmt.Errorf("failed to request,status code:%v", httpResp.StatusCode)
	}

	if resp == nil || resp.BlockchainState.ToPointer() == nil {
		return false, fmt.Errorf("cannot get response from node")
	}

	if resp.Error.ToPointer() != nil {
		return false, fmt.Errorf(*resp.Error.ToPointer())
	}

	syncState := resp.BlockchainState.ToPointer().Sync
	return syncState.Synced, nil
}

func (cli *Client) GetBalance(ctx context.Context, address string) (uint64, error) {
	_, addressPH, err := puzzlehash.GetPuzzleHashFromAddress(address)
	if err != nil || addressPH == nil {
		return 0, fmt.Errorf("invalid address,err: %v", err)
	}

	resp, httpResp, err := cli.fullNodeService.GetCoinRecordsByPuzzleHash(ctx, &GetCoinRecordsByPuzzleHashOptions{
		PuzzleHash:        *addressPH,
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

	if resp.Error.ToPointer() != nil {
		return 0, fmt.Errorf(*resp.Error.ToPointer())
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

func (cli *Client) SelectCoins(ctx context.Context, totalAmount uint64, puzzleHash types.Bytes32) ([]*types.Coin, error) {
	resp, httpResp, err := cli.fullNodeService.GetCoinRecordsByPuzzleHash(ctx, &GetCoinRecordsByPuzzleHashOptions{
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

	if resp.Error.ToPointer() != nil {
		return nil, fmt.Errorf(*resp.Error.ToPointer())
	}

	return selectCoins(totalAmount, resp.CoinRecords)
}

func (cli *Client) PushTX(ctx context.Context, SpendBundle *types.SpendBundle) (string, error) {
	resp, httpResp, err := cli.fullNodeService.PushTX(ctx, &FullNodePushTXOptions{*SpendBundle})
	if err != nil {
		return "", err
	}

	if httpResp.StatusCode != 200 {
		return "", fmt.Errorf("failed to request, status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return "", fmt.Errorf("cannot get response from node")
	}

	if resp.Error.ToPointer() != nil {
		return "", fmt.Errorf(*resp.Error.ToPointer())
	}

	if !resp.Success {
		return "", fmt.Errorf("failed to push tx, status:%v", resp.Status)
	}

	return calTxHash(SpendBundle)
}

func (cli *Client) CheckTxIDInMempool(ctx context.Context, txid string) (bool, error) {
	txidBytes, err := hex.DecodeString(txid)
	if err != nil {
		return false, fmt.Errorf("invalid txid,err: %v", err)
	}

	txidBytes32, err := types.BytesToBytes32(txidBytes)
	if err != nil {
		return false, fmt.Errorf("invalid txid,err: %v", err)
	}

	resp, httpResp, err := cli.fullNodeService.CheckTxIDInMempool(ctx, &CheckTxIDInMempoolOptions{
		TxID: txidBytes32,
	})

	if err != nil {
		return false, err
	}

	if httpResp.StatusCode != 200 {
		return false, fmt.Errorf("failed to request, status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return false, fmt.Errorf("cannot get response from node")
	}

	errMsg := resp.Error.ToPointer()
	if errMsg != nil && strings.Contains(*errMsg, "not in the mempool") {
		return false, nil
	}

	if resp.Error.ToPointer() != nil {
		return false, fmt.Errorf(*resp.Error.ToPointer())
	}

	if !resp.Success {
		return false, nil
	}

	return true, nil
}

// all coin spent return true
func (cli *Client) CheckCoinsIsSpent(ctx context.Context, coinids []string) (bool, error) {
	resp, httpResp, err := cli.fullNodeService.GetCoinRecordsByNames(ctx,
		&GetCoinRecordByNamesOptions{
			Names:             coinids,
			IncludeSpentCoins: true,
		},
	)
	if err != nil {
		return false, err
	}

	if httpResp.StatusCode != 200 {
		return false, fmt.Errorf("failed to request, status code:%v", httpResp.StatusCode)
	}

	if resp == nil {
		return false, fmt.Errorf("cannot get response from node")
	}

	if resp.Error.ToPointer() != nil {
		return false, fmt.Errorf(*resp.Error.ToPointer())
	}

	if !resp.Success {
		return false, fmt.Errorf("failed to query from node")
	}

	if len(resp.CoinRecords) != len(coinids) {
		return false, fmt.Errorf("some records not found")
	}

	for _, record := range resp.CoinRecords {
		if !record.Spent {
			return false, nil
		}
	}

	return true, nil
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

func calTxHash(spendBundele *types.SpendBundle) (string, error) {
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
