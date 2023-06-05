package hc_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/cneill/hc"
)

func TestGetJSON(t *testing.T) {
	t.Parallel()

	opts := hc.DefaultOpts()
	opts.Debug = true
	client := hc.New(opts)
	type dummyStruct struct {
		UserURL string `json:"user_url"`
	}
	ds := &dummyStruct{}
	client.GetJSON("https://api.github.com/", ds)
	fmt.Printf("%#v\n", ds)
}

func TestPrepareURLAppendSlash(t *testing.T) {
	t.Parallel()

	opts := hc.DefaultOpts()
	opts.AppendSlash = true
	client := hc.New(opts)

	tests := []struct {
		inputURL  string
		outputURL string
	}{
		{"https://test.com", "https://test.com/"},
		{"https://test.com/rest", "https://test.com/rest/"},
		{"https://test.com/rest?query=1", "https://test.com/rest/?query=1"},
		{"https://test.com/rest&query=1", "https://test.com/rest/&query=1"},
	}

	for _, test := range tests {
		url, err := client.PrepareURL(test.inputURL)
		if err != nil {
			t.Errorf("failed to prepare URL: %v", err)
		}

		if output := url.String(); output != test.outputURL {
			t.Errorf("expected output URL %q, got %q", test.outputURL, output)
		}
	}
}

func TestPrepareURLAddedQuery(t *testing.T) {
	t.Parallel()

	opts := hc.DefaultOpts()
	client := hc.New(opts)

	tests := []struct {
		inputURL   string
		outputURL  string
		addedQuery url.Values
	}{
		{"https://test.com", "https://test.com/?query1=1&query2=2&query2=3", url.Values{"query1": []string{"1"}, "query2": []string{"2", "3"}}},
		{"https://test.com/path", "https://test.com/path?query1=1&query2=2&query2=3", url.Values{"query1": []string{"1"}, "query2": []string{"2", "3"}}},
		{"https://test.com/path?query1=1", "https://test.com/path?query1=1&query2=2&query2=3", url.Values{"query2": []string{"2", "3"}}},
	}

	for _, test := range tests {
		client.Opts.AddedQuery = test.addedQuery

		url, err := client.PrepareURL(test.inputURL)
		if err != nil {
			t.Errorf("failed to prepare URL: %v", err)
		}

		if output := url.String(); output != test.outputURL {
			t.Errorf("expected output URL %q, got %q", test.outputURL, output)
		}
	}
}

/*
func TestAddQueryValues(t *testing.T) {
	var vals = url.Values{}
	vals.Add("test", "testVal")
	vals.Add("test", "test2")
	vals.Add("test2", "testval2")

	req, err := http.NewRequest(http.MethodGet, "https://google.com/", nil)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	AddQueryValues(req, vals)

	for key, keyVals := range req.URL.Query() {
		comparison := vals[key]

		if len(keyVals) != len(comparison) {
			t.Errorf("wrong number of values, expected %d, got %d", len(comparison), len(keyVals))
		}

		for _, val := range keyVals {
			var found = false
			for _, comparisonVal := range comparison {
				if comparisonVal == val {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("missing value %q for key %q", val, key)
			}
		}
	}
}
*/
