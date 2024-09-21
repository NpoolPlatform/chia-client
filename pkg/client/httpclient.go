package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/rpcinterface"
)

// WalletService encapsulates wallet RPC methods
type HttpClient struct {
	Endpoint    string // host:port
	BasePath    string // http://host:port/basePath/request
	Timeout     time.Duration
	serviceType rpcinterface.ServiceType
}

// NewRequest creates an RPC request for the specified service
func (c *HttpClient) NewRequest(rpcEndpoint rpcinterface.Endpoint, opt interface{}) (*rpcinterface.Request, error) {
	// Always POST
	// Supporting it as a variable in case that changes in the future, it can be passed in instead
	method := http.MethodPost

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Accept", "application/json")

	var body []byte
	var err error
	switch {
	case method == http.MethodPost || method == http.MethodPut:
		reqHeaders.Set("Content-Type", "application/json")

		// Always need at least an empty json object in the body
		if opt == nil {
			body = []byte(`{}`)
		} else {
			body, err = json.Marshal(opt)
			if err != nil {
				return nil, err
			}
		}
	}

	url := fmt.Sprintf("http://%v/%v/%v", c.Endpoint, c.BasePath, rpcEndpoint)

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return &rpcinterface.Request{
		Service: c.serviceType,
		Request: req,
	}, nil
}

// Do sends an RPC request and returns the RPC response.
func (c *HttpClient) Do(req *rpcinterface.Request, v interface{}) (*http.Response, error) {
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
