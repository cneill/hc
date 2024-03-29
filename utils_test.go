package hc_test

import (
	"fmt"
	"testing"

	"github.com/cneill/hc"
)

type testStringer struct{}

func (t testStringer) String() string { return "string" }

type EmbeddedTest struct {
	EmbeddedTest string `url:"embedded_test"`
}

type anonymousTest struct {
	AnonymousTest string `url:"anonymous_test"`
}

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
				OmitEmptyPopulatedTest string `url:"omitempty_populated_test,omitempty"`
			}{"string"},
			"omitempty_populated_test",
			[]string{"string"},
		},
		{
			struct {
				OmitEmptyUnpopulatedTest string `url:"omitempty_unpopulated_test,omitempty"`
			}{},
			"omitempty_unpopulated_test",
			[]string{},
		},
		{
			struct {
				OmitEmptyNilTest *testStringer `url:"omitempty_nil_test,omitempty"`
			}{nil},
			"omitempty_nil_test",
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
		{
			struct {
				unexportedTest string `url:"unexported_test"`
			}{"string"},
			"unexported_test",
			[]string{},
		},
		{
			struct {
				EmbeddedTest
			}{EmbeddedTest{"string"}},
			"embedded_test",
			[]string{"string"},
		},
		{
			struct {
				anonymousTest
			}{anonymousTest{"string"}},
			"anonymous_test",
			[]string{},
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
