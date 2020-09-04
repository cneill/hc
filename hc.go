package hc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Opts sets options for HC
type Opts struct {
	AddedHeaders http.Header
	Debug        bool
	DebugLogger  *log.Logger
}

// DefaultOpts returns a reasonable Opts object for general use
func DefaultOpts() *Opts {
	return &Opts{
		AddedHeaders: http.Header{},
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

// Do executes an HTTP request
func (h *HC) Do(req *http.Request) (*http.Response, error) {
	for header, values := range h.Opts.AddedHeaders {
		for _, val := range values {
			req.Header.Add(header, val)
		}
	}

	if h.Opts.Debug {
		h.Opts.DebugLogger.Printf("%s %s\n", req.Method, req.URL.String())
	}

	return h.Client.Do(req)
}

// DoJSON performs an HTTP request and decodes the response body into the "response" object provided
func (h *HC) DoJSON(request *http.Request, responseObject interface{}) error {
	resp, err := h.Do(request)
	if err != nil {
		return err
	} else if resp.Body == nil {
		// TODO: determine if this is overkill
		return fmt.Errorf("hc: no body returned")
	}

	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	d.Decode(responseObject)

	return nil
}

// GetJSON creates a simple HTTP GET request and uses DoJSON to execute it
func (h *HC) GetJSON(URL string, responseObject interface{}) error {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	return h.DoJSON(req, responseObject)
}

// PostJSON creates an HTTP POST request with Content-Type "application/json" and uses DoJSON to execute it
func (h *HC) PostJSON(URL string, bodyObject, responseObject interface{}) error {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	e := json.NewEncoder(w)
	e.Encode(bodyObject)

	req, err := http.NewRequest(http.MethodPost, URL, r)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	return h.DoJSON(req, responseObject)
}
