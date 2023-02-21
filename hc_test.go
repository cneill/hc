package hc_test

import (
	"fmt"
	"testing"

	"github.com/cneill/hc"
)

func TestGetJSON(t *testing.T) {
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
