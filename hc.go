package hc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func DefaultClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			IdleConnTimeout:       time.Second * 30,
			TLSHandshakeTimeout:   time.Second * 5,
			ExpectContinueTimeout: time.Second * 1,
			ResponseHeaderTimeout: time.Second * 5,
			DialContext: (&net.Dialer{
				Timeout:   time.Second * 5,
				KeepAlive: time.Second * 30,
			}).DialContext,
			MaxIdleConns:      100,
			ForceAttemptHTTP2: true,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					// TLSv1.3 ciphersuites are not configurable, per a comment from the [crypto/tls] package.
					// TLSv1.2 ciphersuites
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("%w: 10", ErrMaxRedirects)
			}

			if len(via) == 0 {
				return nil
			}

			last := via[len(via)-1]

			if last.URL.Scheme == "https" && req.URL.Scheme != "https" {
				return fmt.Errorf("%w: %s to %s", ErrDowngradeRedirect, last.URL.String(), req.URL.String())
			}

			return nil
		},
	}
}

type httpsOnlyRoundTripper struct {
	*http.Transport
}

func paranoidRoundTripper(transport *http.Transport) http.RoundTripper {
	return &httpsOnlyRoundTripper{
		Transport: transport,
	}
}

func (h *httpsOnlyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.EqualFold(req.URL.Scheme, "https") {
		return nil, fmt.Errorf("%w: %s", ErrNonHTTPSConnection, req.URL.String())
	}

	return h.Transport.RoundTrip(req) //nolint:wrapcheck
}

func ParanoidClient() *http.Client {
	client := DefaultClient()

	transport, _ := client.Transport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	transport.ForceAttemptHTTP2 = false // Revisit if getting blocked by scraper-detectors for only supporting 1.1

	protocols := &http.Protocols{}
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(false)
	protocols.SetUnencryptedHTTP2(false)
	transport.Protocols = protocols

	client.Transport = paranoidRoundTripper(transport)
	defaultRedir := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := defaultRedir(req, via); err != nil {
			return err
		}

		if len(via) == 0 {
			return nil
		}

		last := via[len(via)-1]

		// allows redirect from example.com -> www.example.com
		if !strings.EqualFold(last.URL.Hostname(), req.URL.Hostname()) &&
			!strings.EqualFold("www."+last.URL.Hostname(), req.URL.Hostname()) {
			return fmt.Errorf("%w: %s to %s", ErrCrossDomainRedirect, last.URL.String(), req.URL.String())
		}

		if last.URL.Scheme != req.URL.Scheme {
			return fmt.Errorf("%w: %s to %s", ErrSchemeSwitchRedirect, last.URL.String(), req.URL.String())
		}

		normalizePort := func(reqURL *url.URL) (int, error) {
			port := reqURL.Port()
			switch {
			case port == "" && strings.EqualFold(reqURL.Scheme, "https"):
				return 443, nil
			case port == "" && strings.EqualFold(reqURL.Scheme, "http"):
				return 80, nil
			default:
				intPort, err := strconv.Atoi(port)
				if err != nil {
					return 0, fmt.Errorf("%w: %s", ErrInvalidPort, port)
				}

				return intPort, nil
			}
		}

		lastPort, err := normalizePort(last.URL)
		if err != nil {
			return err
		}

		nextPort, err := normalizePort(req.URL)
		if err != nil {
			return err
		}

		if lastPort != nextPort {
			return fmt.Errorf("%w: %s to %s", ErrPortMismatch, last.URL.String(), req.URL.String())
		}

		return nil
	}

	return client
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

type (
	ClientOpt  func(hc *HC) (*HC, error)
	ClientOpts []ClientOpt
)

type HC struct {
	client       *http.Client
	rateLimiter  *time.Ticker
	addedHeaders http.Header
	baseURL      string
}

func New(client *http.Client, opts ...ClientOpt) (*HC, error) {
	hc := &HC{
		client: client,
	}

	var optErr error
	for i, opt := range opts {
		hc, optErr = opt(hc)
		if optErr != nil {
			return nil, fmt.Errorf("opt %d: %w", i+1, optErr)
		}
	}

	return hc, nil
}

func (h *HC) Do(opts ...RequestOpt) (*http.Response, error) {
	if h.rateLimiter != nil {
		<-h.rateLimiter.C
	}

	allOpts := RequestOpts{}

	if h.addedHeaders != nil {
		for name, vals := range h.addedHeaders {
			for _, val := range vals {
				allOpts = append(allOpts, Header(name, val))
			}
		}
	}

	if h.baseURL != "" {
		allOpts = append(allOpts, BaseURL(h.baseURL))
	}

	allOpts = append(allOpts, opts...)

	req, err := NewRequest(allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return h.client.Do(req)
}

func ClientBaseURL(base string) ClientOpt {
	return func(hc *HC) (*HC, error) {
		_, err := url.Parse(base)
		if err != nil {
			return nil, fmt.Errorf("ClientBaseURL: failed to parse URL: %w", err)
		}

		hc.baseURL = base

		return hc, nil
	}
}

func RateTickerDuration(duration time.Duration) ClientOpt {
	return func(hc *HC) (*HC, error) {
		hc.rateLimiter = time.NewTicker(duration)
		return hc, nil
	}
}

func GlobalHeaders(headers http.Header) ClientOpt {
	return func(hc *HC) (*HC, error) {
		hc.addedHeaders = headers
		return hc, nil
	}
}
