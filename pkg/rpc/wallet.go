package rpc

import (
	"net/url"

	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	rpc1 "github.com/chia-network/go-chia-libs/pkg/rpc"
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type MyWalletService struct {
	*rpc1.WalletService
}

type WalletOption struct {
	WalletID uint32 `json:"wallet_id"`
	Amount   uint64 `json:"amount"`
}

type SelectCoinsResponse struct {
	rpc1.Response
	Coins []types1.Coin `json:"coins"`
}

func (s *MyWalletService) SelectCoins(opts *WalletOption) (*[]types1.Coin, error) {
	request, err := s.NewRequest("select_coins", opts)
	if err != nil {
		return nil, err
	}

	r := &SelectCoinsResponse{}
	_, err = s.Do(request, r)
	if err != nil {
		return nil, err
	}
	return &r.Coins, nil
}

type GetWalletBalanceOptions struct {
	WalletID uint32 `json:"wallet_id"`
}

type GetWalletBalanceResponse struct {
	rpc1.Response
	Balance types1.WalletBalance `json:"wallet_balance"`
}

func (s *MyWalletService) GetWalletBalance(opts *GetWalletBalanceOptions) (*types1.WalletBalance, error) {
	request, err := s.NewRequest("get_wallet_balance", opts)
	if err != nil {
		return nil, err
	}

	r := &GetWalletBalanceResponse{}
	_, err = s.Do(request, r)
	if err != nil {
		return nil, err
	}
	return &r.Balance, nil
}

func GetWalletClient() (*MyWalletService, error) {
	client, err := rpc1.NewClient(rpc1.ConnectionModeHTTP, rpc1.WithAutoConfig(), rpc1.WithBaseURL(&url.URL{Scheme: "https", Host: "localhost"}))
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	walletService := &MyWalletService{
		WalletService: client.WalletService,
	}
	return walletService, nil
}
