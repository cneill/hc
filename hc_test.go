package hc_test

import (
	"strings"
	"testing"

	"github.com/cneill/hc/v2"
)

func TestClientConnections(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		paranoid    bool
		wantErrText string
	}{
		{
			name:        "Google",
			url:         "https://www.google.com/",
			paranoid:    false,
			wantErrText: "",
		},
		{
			name:        "Paranoid Google redirect to www",
			url:         "https://google.com/",
			paranoid:    true,
			wantErrText: "",
		},
		{
			name:        "icanhazip",
			url:         "https://icanhazip.com/",
			paranoid:    false,
			wantErrText: "",
		},
		{
			name:        "Paranoid icanhazip",
			url:         "https://icanhazip.com/",
			paranoid:    true,
			wantErrText: "",
		},
		{
			name:        "BadSSL TLSv1.1",
			url:         "https://tls-v1-1.badssl.com:1011/",
			paranoid:    false,
			wantErrText: "tls: handshake failure",
		},
		{
			name:        "BadSSL HTTP",
			url:         "http://http.badssl.com/",
			paranoid:    false,
			wantErrText: "",
		},
		{
			name:        "Paranoid BadSSL HTTP",
			url:         "http://http.badssl.com/",
			paranoid:    true,
			wantErrText: "paranoid: rejecting non-HTTPS connection",
		},
		{
			name:        "BadSSL RC4",
			url:         "https://rc4.badssl.com/",
			paranoid:    false,
			wantErrText: "tls: handshake failure",
		},
		{
			name:        "BadSSL TLSv1.2",
			url:         "https://tls-v1-2.badssl.com:1012/",
			paranoid:    false,
			wantErrText: "",
		},
		{
			name:        "Paranoid BadSSL TLSv1.2",
			url:         "https://tls-v1-2.badssl.com:1012/",
			paranoid:    true,
			wantErrText: "tls: handshake failure",
		},
	}

	for _, test := range tests {
		defaultClient := hc.DefaultClient()
		paranoidClient := hc.ParanoidClient()

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var err error

			if test.paranoid {
				_, err = paranoidClient.Get(test.url)
			} else {
				_, err = defaultClient.Get(test.url)
			}

			if test.wantErrText != "" {
				if err == nil {
					t.Fatalf("expecting error text %q requesting URL %q, got nil", test.wantErrText, test.url)
				}

				if errText := err.Error(); !strings.Contains(errText, test.wantErrText) {
					t.Fatalf("expecting error text %q requesting URL %q, got %q", test.wantErrText, test.url, errText)
				}
			} else if err != nil {
				t.Fatalf("expecting nil error requesting URL %q, got %v", test.url, err)
			}
		})
	}
}
