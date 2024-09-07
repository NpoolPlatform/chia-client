package rpc

import (
	"net/url"

	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	rpc1 "github.com/chia-network/go-chia-libs/pkg/rpc"
	types1 "github.com/chia-network/go-chia-libs/pkg/types"
)

type MyFullNodeService struct {
	*rpc1.FullNodeService
}

func (s *MyFullNodeService) GetCoinsByPuzzleHash(opts *rpc1.GetCoinRecordsByPuzzleHashOptions) ([]types1.CoinRecord, error) {
	resp, _, err := s.GetCoinRecordsByPuzzleHash(opts)
	if err != nil {
		return nil, err
	}
	return resp.CoinRecords, nil
}

type GetAggSigAdditionalDataResponse struct {
	AdditionalData string `json:"additional_data"`
}

func (s *MyFullNodeService) GetAggSigAdditionalData() (*string, error) {
	request, err := s.NewRequest("get_aggsig_additional_data", nil)
	if err != nil {
		return nil, err
	}

	r := &GetAggSigAdditionalDataResponse{}
	_, err = s.Do(request, r)
	if err != nil {
		return nil, err
	}

	return &r.AdditionalData, nil
}

func GetFullNodeClient() (*MyFullNodeService, error) {
	client, err := rpc1.NewClient(rpc1.ConnectionModeHTTP, rpc1.WithAutoConfig(), rpc1.WithBaseURL(&url.URL{Scheme: "https", Host: "localhost"}))
	if err != nil {
		return nil, wlog.WrapError(err)
	}
	fullNodeService := &MyFullNodeService{
		FullNodeService: client.FullNodeService,
	}
	return fullNodeService, nil
}
