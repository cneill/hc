package hc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

func DefaultClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
				CipherSuites: []uint16{
					// TLSv1.3
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,

					// TLSv1.2
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				},
			},
		},
	}
}

type Option interface {
	RequestOpt | RequestOpts |
		ResponseOpt | ResponseOpts
}

func NewRequest(opts ...RequestOpt) (*http.Request, error) {
	req := &http.Request{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	req = req.WithContext(context.Background())

	var optErr error

	for i, opt := range opts {
		req, optErr = opt(req)
		if optErr != nil {
			return nil, fmt.Errorf("opt %d: %w", i+1, optErr)
		}
	}

	return req, nil
}

func HandleResponse(response *http.Response, opts ...ResponseOpt) (*http.Response, error) {
	var optErr error
	for i, opt := range opts {
		response, optErr = opt(response)
		if optErr != nil {
			return nil, fmt.Errorf("opt %d: %w", i+1, optErr)
		}
	}

	return response, nil
}

func Do(client *http.Client, opts ...RequestOpt) (*http.Response, error) {
	req, err := NewRequest(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to produce request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}
