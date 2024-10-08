package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/samber/mo"

	"github.com/chia-network/go-chia-libs/pkg/rpcinterface"
	"github.com/chia-network/go-chia-libs/pkg/types"
)

type WalletService struct {
	*HttpClient
}

func DefaultWalletService(endpoint string) *WalletService {
	return &WalletService{
		HttpClient: &HttpClient{
			Endpoint:    endpoint,
			BasePath:    DefaultBasePath,
			Timeout:     DefaultTimeout,
			serviceType: rpcinterface.ServiceFullNode,
		},
	}
}

// Do sends an RPC request and returns the RPC response.
func (c *WalletService) Do(req *rpcinterface.Request, v interface{}) (*http.Response, error) {
	client := http.Client{
		Timeout: c.Timeout,
	}

	resp, err := client.Do(req.Request)
	if err != nil {
		return nil, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return resp, err
}

// GetConnections returns connections
func (s *WalletService) GetConnections(ctx context.Context, opts *GetConnectionsOptions) (*GetConnectionsResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_connections", opts)
	if err != nil {
		return nil, nil, err
	}

	c := &GetConnectionsResponse{}
	resp, err := s.Do(request, c)
	if err != nil {
		return nil, resp, err
	}

	return c, resp, nil
}

// GetNetworkInfo wallet rpc -> get_network_info
func (s *WalletService) GetNetworkInfo(ctx context.Context, opts *GetNetworkInfoOptions) (*GetNetworkInfoResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_network_info", nil)
	if err != nil {
		return nil, nil, err
	}

	r := &GetNetworkInfoResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetVersion returns the application version for the service
func (s *WalletService) GetVersion(ctx context.Context, opts *GetVersionOptions) (*GetVersionResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_version", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetVersionResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetPublicKeysResponse response from get_public_keys
type GetPublicKeysResponse struct {
	Response
	PublicKeyFingerprints mo.Option[[]int] `json:"public_key_fingerprints"`
}

// GetPublicKeys endpoint
func (s *WalletService) GetPublicKeys(ctx context.Context) (*GetPublicKeysResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_public_keys", nil)
	if err != nil {
		return nil, nil, err
	}

	r := &GetPublicKeysResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GenerateMnemonicResponse Random new 24 words response
type GenerateMnemonicResponse struct {
	Response
	Mnemonic mo.Option[[]string] `json:"mnemonic"`
}

// GenerateMnemonic Endpoint for generating a new random 24 words
func (s *WalletService) GenerateMnemonic(ctx context.Context) (*GenerateMnemonicResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "generate_mnemonic", nil)
	if err != nil {
		return nil, nil, err
	}

	r := &GenerateMnemonicResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// AddKeyOptions options for the add_key endpoint
type AddKeyOptions struct {
	Mnemonic []string `json:"mnemonic"`
}

// AddKeyResponse response from the add_key endpoint
type AddKeyResponse struct {
	Response
	Word        mo.Option[string] `json:"word,omitempty"` // This is part of a unique error response
	Fingerprint mo.Option[int]    `json:"fingerprint,omitempty"`
}

// AddKey Adds a new key from 24 words to the keychain
func (s *WalletService) AddKey(ctx context.Context, options *AddKeyOptions) (*AddKeyResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "add_key", options)
	if err != nil {
		return nil, nil, err
	}

	r := &AddKeyResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// DeleteAllKeysResponse Delete keys response
type DeleteAllKeysResponse struct {
	Response
}

// DeleteAllKeys deletes all keys from the keychain
func (s *WalletService) DeleteAllKeys(ctx context.Context) (*DeleteAllKeysResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "delete_all_keys", nil)
	if err != nil {
		return nil, nil, err
	}

	r := &DeleteAllKeysResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetNextAddressOptions options for get_next_address endpoint
type GetNextAddressOptions struct {
	NewAddress bool   `json:"new_address"`
	WalletID   uint32 `json:"wallet_id"`
}

// GetNextAddressResponse response from get next address
type GetNextAddressResponse struct {
	Response
	WalletID mo.Option[uint32] `json:"wallet_id"`
	Address  mo.Option[string] `json:"address"`
}

// GetNextAddress returns the current address for the wallet. If NewAddress is true, it moves to the next address before responding
func (s *WalletService) GetNextAddress(ctx context.Context, options *GetNextAddressOptions) (*GetNextAddressResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_next_address", options)
	if err != nil {
		return nil, nil, err
	}

	r := &GetNextAddressResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletSyncStatusResponse Response for get_sync_status on wallet
type GetWalletSyncStatusResponse struct {
	Response
	GenesisInitialized mo.Option[bool] `json:"genesis_initialized"`
	Synced             mo.Option[bool] `json:"synced"`
	Syncing            mo.Option[bool] `json:"syncing"`
}

// GetWalletHeightInfoResponse response for get_height_info on wallet
type GetWalletHeightInfoResponse struct {
	Response
	Height mo.Option[uint32] `json:"height"`
}

// GetHeightInfo wallet rpc -> get_height_info
func (s *WalletService) GetHeightInfo(ctx context.Context) (*GetWalletHeightInfoResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_height_info", nil)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletHeightInfoResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletsOptions wallet rpc -> get_wallets
type GetWalletsOptions struct {
	Type types.WalletType `json:"type"`
}

// GetWalletsResponse wallet rpc -> get_wallets
type GetWalletsResponse struct {
	Response
	Fingerprint mo.Option[int]                `json:"fingerprint"`
	Wallets     mo.Option[[]types.WalletInfo] `json:"wallets"`
}

// GetWallets wallet rpc -> get_wallets
func (s *WalletService) GetWallets(ctx context.Context, opts *GetWalletsOptions) (*GetWalletsResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_wallets", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletsResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletBalanceOptions request options for get_wallet_balance
type GetWalletBalanceOptions struct {
	WalletID uint32 `json:"wallet_id"`
}

// GetWalletBalanceResponse is the wallet balance RPC response
type GetWalletBalanceResponse struct {
	Response
	Balance mo.Option[types.WalletBalance] `json:"wallet_balance"`
}

// GetWalletBalance returns wallet balance
func (s *WalletService) GetWalletBalance(ctx context.Context, opts *GetWalletBalanceOptions) (*GetWalletBalanceResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_wallet_balance", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletBalanceResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletTransactionCountOptions options for get transaction count
type GetWalletTransactionCountOptions struct {
	WalletID uint32 `json:"wallet_id"`
}

// GetWalletTransactionCountResponse response for get_transaction_count
type GetWalletTransactionCountResponse struct {
	Response
	WalletID mo.Option[uint32] `json:"wallet_id"`
	Count    mo.Option[int]    `json:"count"`
}

// GetTransactionCount returns the total count of transactions for the specific wallet ID
func (s *WalletService) GetTransactionCount(ctx context.Context, opts *GetWalletTransactionCountOptions) (*GetWalletTransactionCountResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_transaction_count", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletTransactionCountResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletTransactionsOptions options for get wallet transactions
type GetWalletTransactionsOptions struct {
	WalletID  uint32 `json:"wallet_id"`
	Start     *int   `json:"start,omitempty"`
	End       *int   `json:"end,omitempty"`
	ToAddress string `json:"to_address,omitempty"`
}

// GetWalletTransactionsResponse response for get_wallet_transactions
type GetWalletTransactionsResponse struct {
	Response
	WalletID     mo.Option[uint32]                    `json:"wallet_id"`
	Transactions mo.Option[[]types.TransactionRecord] `json:"transactions"`
}

// GetTransactions wallet rpc -> get_transactions
func (s *WalletService) GetTransactions(ctx context.Context, opts *GetWalletTransactionsOptions) (*GetWalletTransactionsResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_transactions", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletTransactionsResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetWalletTransactionOptions options for getting a single wallet transaction
type GetWalletTransactionOptions struct {
	WalletID      uint32 `json:"wallet_id"`
	TransactionID string `json:"transaction_id"`
}

// GetWalletTransactionResponse response for get_wallet_transactions
type GetWalletTransactionResponse struct {
	Response
	Transaction   mo.Option[types.TransactionRecord] `json:"transaction"`
	TransactionID mo.Option[string]                  `json:"transaction_id"`
}

// GetTransaction returns a single transaction record
func (s *WalletService) GetTransaction(ctx context.Context, opts *GetWalletTransactionOptions) (*GetWalletTransactionResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_transaction", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetWalletTransactionResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// SendTransactionOptions represents the options for send_transaction
type SendTransactionOptions struct {
	WalletID uint32        `json:"wallet_id"`
	Amount   uint64        `json:"amount"`
	Address  string        `json:"address"`
	Memos    []types.Bytes `json:"memos,omitempty"`
	Fee      uint64        `json:"fee"`
	Coins    []types.Coin  `json:"coins,omitempty"`
}

// SendTransactionResponse represents the response from send_transaction
type SendTransactionResponse struct {
	Response
	TransactionID mo.Option[string]                  `json:"transaction_id"`
	Transaction   mo.Option[types.TransactionRecord] `json:"transaction"`
}

// SendTransaction sends a transaction
func (s *WalletService) SendTransaction(ctx context.Context, opts *SendTransactionOptions) (*SendTransactionResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "send_transaction", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &SendTransactionResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// CatSpendOptions represents the options for cat_spend
type CatSpendOptions struct {
	WalletID uint32 `json:"wallet_id"`
	Amount   uint64 `json:"amount"`
	Address  string `json:"inner_address"`
	Fee      uint64 `json:"fee"`
}

// CatSpendResponse represents the response from cat_spend
type CatSpendResponse struct {
	Response
	TransactionID mo.Option[string]                  `json:"transaction_id"`
	Transaction   mo.Option[types.TransactionRecord] `json:"transaction"`
}

// CatSpend sends a transaction
func (s *WalletService) CatSpend(ctx context.Context, opts *CatSpendOptions) (*CatSpendResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "cat_spend", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &CatSpendResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// MintNFTOptions represents the options for nft_get_info
type MintNFTOptions struct {
	DidID             string   `json:"did_id"`             // not required
	EditionNumber     uint32   `json:"edition_number"`     // not required
	EditionCount      uint32   `json:"edition_count"`      // not required
	Fee               uint64   `json:"fee"`                // not required
	LicenseHash       string   `json:"license_hash"`       //not required
	LicenseURIs       []string `json:"license_uris"`       // not required
	MetaHash          string   `json:"meta_hash"`          // not required
	MetaURIs          []string `json:"meta_uris"`          // not required
	RoyaltyAddress    string   `json:"royalty_address"`    // not required
	RoyaltyPercentage uint32   `json:"royalty_percentage"` // not required
	TargetAddress     string   `json:"target_address"`     // not required
	Hash              string   `json:"hash"`
	URIs              []string `json:"uris"`
	WalletID          uint32   `json:"wallet_id"`
}

// MintNFTResponse represents the response from nft_get_info
type MintNFTResponse struct {
	Response
	SpendBundle mo.Option[types.SpendBundle] `json:"spend_bundle"`
	WalletID    mo.Option[uint32]            `json:"wallet_id"`
}

// MintNFT Mint a new NFT
func (s *WalletService) MintNFT(ctx context.Context, opts *MintNFTOptions) (*MintNFTResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_mint_nft", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &MintNFTResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetNFTsOptions represents the options for nft_get_nfts
type GetNFTsOptions struct {
	WalletID   uint32         `json:"wallet_id"`
	StartIndex mo.Option[int] `json:"start_index"`
	Num        mo.Option[int] `json:"num"`
}

// GetNFTsResponse represents the response from nft_get_nfts
type GetNFTsResponse struct {
	Response
	WalletID mo.Option[uint32]          `json:"wallet_id"`
	NFTList  mo.Option[[]types.NFTInfo] `json:"nft_list"`
}

// GetNFTs Show all NFTs in a given wallet
func (s *WalletService) GetNFTs(ctx context.Context, opts *GetNFTsOptions) (*GetNFTsResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_get_nfts", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetNFTsResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// TransferNFTOptions represents the options for nft_get_info
type TransferNFTOptions struct {
	Fee           uint64 `json:"fee"` // not required
	NFTCoinID     string `json:"nft_coin_id"`
	TargetAddress string `json:"target_address"`
	WalletID      uint32 `json:"wallet_id"`
}

// TransferNFTResponse represents the response from nft_get_info
type TransferNFTResponse struct {
	Response
	SpendBundle mo.Option[types.SpendBundle] `json:"spend_bundle"`
	WalletID    mo.Option[uint32]            `json:"wallet_id"`
}

// TransferNFT Get info about an NFT
func (s *WalletService) TransferNFT(ctx context.Context, opts *TransferNFTOptions) (*TransferNFTResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_transfer_nft", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &TransferNFTResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetNFTInfoOptions represents the options for nft_get_info
type GetNFTInfoOptions struct {
	CoinID   string `json:"coin_id"`
	WalletID uint32 `json:"wallet_id"`
}

// GetNFTInfoResponse represents the response from nft_get_info
type GetNFTInfoResponse struct {
	Response
	NFTInfo mo.Option[types.NFTInfo] `json:"nft_info"`
}

// GetNFTInfo Get info about an NFT
func (s *WalletService) GetNFTInfo(ctx context.Context, opts *GetNFTInfoOptions) (*GetNFTInfoResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_get_info", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetNFTInfoResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// NFTAddURIOptions represents the options for nft_add_uri
type NFTAddURIOptions struct {
	Fee       uint64 `json:"fee"` // not required
	Key       string `json:"key"`
	NFTCoinID string `json:"nft_coin_id"`
	URI       string `json:"uri"`
	WalletID  uint32 `json:"wallet_id"`
}

// NFTAddURIResponse represents the response from nft_add_uri
type NFTAddURIResponse struct {
	Response
	SpendBundle mo.Option[types.SpendBundle] `json:"spend_bundle"`
	WalletID    mo.Option[uint32]            `json:"wallet_id"`
}

// NFTAddURI Get info about an NFT
func (s *WalletService) NFTAddURI(ctx context.Context, opts *NFTAddURIOptions) (*NFTAddURIResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_add_uri", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &NFTAddURIResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// NFTGetByDidOptions represents the options for nft_get_by_did
type NFTGetByDidOptions struct {
	DidID types.Bytes32 `json:"did_id,omitempty"`
}

// NFTGetByDidResponse represents the response from nft_get_by_did
type NFTGetByDidResponse struct {
	Response
	WalletID mo.Option[uint32] `json:"wallet_id"`
}

// NFTGetByDid Get wallet ID by DID
func (s *WalletService) NFTGetByDid(ctx context.Context, opts *NFTGetByDidOptions) (*NFTGetByDidResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "nft_get_by_did", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &NFTGetByDidResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// GetSpendableCoinsOptions Options for get_spendable_coins
type GetSpendableCoinsOptions struct {
	WalletID            uint32   `json:"wallet_id"`
	MinCoinAmount       *uint64  `json:"min_coin_amount,omitempty"`
	MaxCoinAmount       *uint64  `json:"max_coin_amount,omitempty"`
	ExcludedCoinAmounts []uint64 `json:"excluded_coin_amounts,omitempty"`
}

// GetSpendableCoinsResponse response from get_spendable_coins
type GetSpendableCoinsResponse struct {
	Response
	ConfirmedRecords     mo.Option[[]types.CoinRecord] `json:"confirmed_records"`
	UnconfirmedRemovals  mo.Option[[]types.CoinRecord] `json:"unconfirmed_removals"`
	UnconfirmedAdditions mo.Option[[]types.CoinRecord] `json:"unconfirmed_additions"`
}

// GetSpendableCoins returns information about the coins in the wallet
func (s *WalletService) GetSpendableCoins(ctx context.Context, opts *GetSpendableCoinsOptions) (*GetSpendableCoinsResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "get_spendable_coins", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &GetSpendableCoinsResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// CreateSignedTransactionOptions Options for create_signed_transaction endpoint
type CreateSignedTransactionOptions struct {
	WalletID           *uint32          `json:"wallet_id,omitempty"`
	Additions          []types.Addition `json:"additions"`
	Fee                *uint64          `json:"fee,omitempty"`
	MinCoinAmount      *uint64          `json:"min_coin_amount,omitempty"`
	MaxCoinAmount      *uint64          `json:"max_coin_amount,omitempty"`
	ExcludeCoinAmounts []*uint64        `json:"exclude_coin_amounts,omitempty"`
	Coins              []types.Coin     `json:"Coins,omitempty"`
	ExcludeCoins       []types.Coin     `json:"exclude_coins,omitempty"`
}

// CreateSignedTransactionResponse Response from create_signed_transaction
type CreateSignedTransactionResponse struct {
	Response
	SignedTXs mo.Option[[]types.TransactionRecord] `json:"signed_txs"`
	SignedTX  mo.Option[types.TransactionRecord]   `json:"signed_tx"`
}

// CreateSignedTransaction generates a signed transaction based on the specified options
func (s *WalletService) CreateSignedTransaction(ctx context.Context, opts *CreateSignedTransactionOptions) (*CreateSignedTransactionResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "create_signed_transaction", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &CreateSignedTransactionResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// SendTransactionMultiResponse Response from send_transaction_multi
type SendTransactionMultiResponse struct {
	Response
	Transaction   mo.Option[types.TransactionRecord] `json:"transaction"`
	TransactionID mo.Option[string]                  `json:"transaction_id"`
}

// SendTransactionMulti allows sending a more detailed transaction with multiple inputs/outputs.
// Options are the same as create signed transaction since this is ultimately just a wrapper around that in Chia
func (s *WalletService) SendTransactionMulti(ctx context.Context, opts *CreateSignedTransactionOptions) (*SendTransactionMultiResponse, *http.Response, error) {
	request, err := s.NewRequest(ctx, "send_transaction_multi", opts)
	if err != nil {
		return nil, nil, err
	}

	r := &SendTransactionMultiResponse{}
	resp, err := s.Do(request, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}
