package client

import (
	"time"

	"github.com/samber/mo"
)

const (
	DefaultBasePath = "fullnode"
	DefaultTimeout  = time.Second * 3
)

// Response is the base response elements that may be in any response from an RPC server in Chia
type Response struct {
	Success bool              `json:"success"`
	Error   mo.Option[string] `json:"error,omitempty"`
}

// GetNetworkInfoOptions options for the get_network_info rpc calls
type GetNetworkInfoOptions struct{}

// GetNetworkInfoResponse common get_network_info response from all RPC services
type GetNetworkInfoResponse struct {
	Response
	NetworkName   mo.Option[string] `json:"network_name"`
	NetworkPrefix mo.Option[string] `json:"network_prefix"`
}

// GetVersionOptions options for the get_version rpc calls
type GetVersionOptions struct{}

// GetVersionResponse is the response of get_version from all RPC services
type GetVersionResponse struct {
	Response
	Version string `json:"version"`
}
