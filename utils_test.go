package hc_test

import (
	"fmt"
	"testing"

	"github.com/cneill/hc"
)

type testStringer struct{}

func (t testStringer) String() string { return "string" }

func TestURLValuesFromStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    any      // struct with variable definitions
		name     string   // name of the url.Values key to check
		expected []string // expected string values for that key
	}{
		{
			struct {
				StringTest string `url:"string_test"`
			}{"string"},
			"string_test",
			[]string{"string"},
		},
		{
			struct {
				IntTest int `url:"int_test"`
			}{123},
			"int_test",
			[]string{"123"},
		},
		{
			struct {
				BoolTest bool `url:"bool_test"`
			}{false},
			"bool_test",
			[]string{"false"},
		},
		{
			struct {
				OmitNameTest string `url:"-"`
			}{"string"},
			"omit_name_test",
			[]string{},
		},
		{
			struct {
				OmitEmptyTest string `url:"omit_empty_test,omitempty"`
			}{"string"},
			"omit_empty_test",
			[]string{"string"},
		},
		{
			struct {
				OmitEmptyTest string `url:"omit_empty_test2,omitempty"`
			}{},
			"omit_empty_test2",
			[]string{},
		},
		{
			struct {
				MultiSliceTest []string `url:"multi_slice_test"`
			}{[]string{"1", "2", "3"}},
			"multi_slice_test",
			[]string{"1", "2", "3"},
		},
		{
			struct {
				ValSepSliceTest []string `url:"val_sep_slice_test" url_val_sep:";"`
			}{[]string{"1", "2", "3"}},
			"val_sep_slice_test",
			[]string{"1;2;3"},
		},
		{
			struct {
				StringerTest testStringer `url:"stringer_test"`
			}{testStringer{}},
			"stringer_test",
			[]string{"string"},
		},
		{
			struct {
				StringerPointerTest *testStringer `url:"stringer_pointer_test"`
			}{&testStringer{}},
			"stringer_pointer_test",
			[]string{"string"},
		},
		/*
			TODO?
			{
				struct {
					MultiArrayTest []string `url:"multi_array_test"`
				}{[3]string{"1", "2", "3"}},
				"multi_slice_test",
				[]string{"1", "2", "3"},
			},
		*/
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			values, err := hc.URLValuesFromStruct(test.input)
			if err != nil {
				t.Fatalf("failed to get values for struct: %v", err)
			}

			thisVal, ok := values[test.name]
			if !ok {
				if n := len(test.expected); n > 0 {
					t.Errorf("no values for that key, expected %d", n)
				}
			} else {
				if len(thisVal) != len(test.expected) {
					t.Errorf("expected %d values, got %d", len(test.expected), len(values))
				}

				for i := 0; i < len(test.expected); i++ {
					if thisVal[i] != test.expected[i] {
						t.Errorf("expected value %d to be %s, got %s", i, test.expected[i], thisVal[i])
					}
				}
			}
		})
	}
}
