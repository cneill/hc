package hc

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Opts sets options for HC
type Opts struct {
	AddedHeaders  http.Header
	AddedQuery    url.Values
	AppendSlash   bool
	BasicAuthUser string
	BasicAuthPass string
	Debug         bool
	DebugLogger   *log.Logger
}

// DefaultOpts returns a reasonable Opts object for general use
func DefaultOpts() *Opts {
	return &Opts{
		AddedHeaders: http.Header{},
		AddedQuery:   url.Values{},
		Debug:        false,
		DebugLogger:  log.New(os.Stderr, "[hc] ", log.Ldate|log.Ltime),
	}
}

// HC is an HTTP client with some helpful features
type HC struct {
	Opts   *Opts
	Client *http.Client
}

// New returns an initialized HC object
func New(opts *Opts) *HC {
	c := &http.Client{
		Transport: &http.Transport{
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
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				},
			},
		},
	}

	return &HC{
		Opts:   opts,
		Client: c,
	}
}

// PrepareURL adds 'values' to the query string of 'URL', appends a "/" if requested, returns error if invalid.
func (h *HC) PrepareURL(URL string) (*url.URL, error) {
	// if requested, append a "/" to the URL if one isn't already present
	if h.Opts.AppendSlash {
		if idx := strings.IndexAny(URL, "?&"); idx != -1 {
			if URL[idx-1] != '/' {
				URL = URL[0:idx] + "/" + URL[idx:]
			}
		} else if !strings.HasSuffix(URL, "/") {
			URL += "/"
		}
	}

	parsed, err := url.Parse(URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provided URL %q: %w", URL, err)
	}

	// append query variables from options to URL
	if len(h.Opts.AddedQuery) > 0 {
		query := parsed.Query()
		for key, vals := range h.Opts.AddedQuery {
			for _, val := range vals {
				query.Add(key, val)
			}
		}

		parsed.RawQuery = query.Encode()

		if parsed.Path == "" {
			parsed.Path = "/"
		}
	}

	return parsed, nil
}

// PrepareRequest applies the various options configured on HC to the provided 'req'.
func (h *HC) PrepareRequest(req *http.Request) (*http.Request, error) {
	// add headers from options
	for header, values := range h.Opts.AddedHeaders {
		for _, val := range values {
			req.Header.Add(header, val)
		}
	}

	newURL, err := h.PrepareURL(req.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare URL to execute HTTP Request: %w", err)
	}

	req.URL = newURL

	return req, nil
}

func (h *HC) simpleRequest(method, URL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}

	if h.Opts.BasicAuthUser != "" && h.Opts.BasicAuthPass != "" {
		req.SetBasicAuth(h.Opts.BasicAuthUser, h.Opts.BasicAuthPass)
	}

	return h.Do(req)
}

func (h *HC) Delete(URL string) (*http.Response, error) {
	return h.simpleRequest(http.MethodDelete, URL, nil)
}

func (h *HC) Get(URL string) (*http.Response, error) {
	return h.simpleRequest(http.MethodGet, URL, nil)
}

func (h *HC) Post(URL string, body io.Reader) (*http.Response, error) {
	return h.simpleRequest(http.MethodPost, URL, body)
}

func (h *HC) Put(URL string, body io.Reader) (*http.Response, error) {
	return h.simpleRequest(http.MethodPut, URL, body)
}

// Do executes an HTTP request
func (h *HC) Do(req *http.Request) (*http.Response, error) {
	req, err := h.PrepareRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	if h.Opts.Debug {
		h.Opts.DebugLogger.Printf("%s %s\n", req.Method, req.URL.String())
	}

	return h.Client.Do(req)
}

// DoJSON performs an HTTP request and decodes the response body into the "response" object provided
func (h *HC) DoJSON(req *http.Request, response any) error {
	resp, err := h.Do(req)
	if err != nil {
		return err
	} else if resp.Body == nil {
		// TODO: determine if this is overkill
		return fmt.Errorf("hc: no body returned")
	}

	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	if err := d.Decode(response); err != nil {
		return fmt.Errorf("hc: failed to decode response object: %w", err)
	}

	return nil
}

// GetJSON creates a simple HTTP GET request and uses DoJSON to execute it
func (h *HC) GetJSON(URL string, response any) error {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	return h.DoJSON(req, response)
}

// PostJSON creates an HTTP POST request with Content-Type "application/json" and uses DoJSON to execute it
func (h *HC) PostJSON(URL string, body, response any) error {
	bodyContents, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("hc: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(bodyContents))
	if err != nil {
		return fmt.Errorf("hc: failed to create request object: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	return h.DoJSON(req, response)
}

// GetStream can be asynchronously called with a goroutine to get events from a "text/event-stream" endpoint delivered on a channel
func (h *HC) GetStream(URL string, events chan string) error {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "text/event-stream")

	resp, err := h.Do(req)
	if err != nil {
		return err
	}

	go h.readStream(resp.Body, events)

	return nil
}

func (h *HC) readStream(stream io.ReadCloser, events chan string) {
	defer stream.Close()

	r := bufio.NewReader(stream)
	for {
		select {
		case <-events:
			break
		default:
			line, err := r.ReadBytes('\n')
			if err == io.EOF {
				close(events)
				break
			}

			if str := strings.TrimSpace(string(line)); str != "" {
				events <- str
			}
		}
	}
}
