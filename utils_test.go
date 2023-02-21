package hc_test

import (
	"testing"

	"github.com/cneill/hc"
)

type testStruct1 struct {
	Name             string    `url:"name"`
	ID               int       `url:"id"`
	Populated        bool      `url:"populated"`
	Omitted          string    `url:"-"`
	OmittedSometimes string    `url:"sometimes,omitempty"`
	MultiSlice       []string  `url:"multi_slice"`
	MultiArray       [3]string `url:"multi_array"`
	SepSlice         []string  `url:"sep_slice" url_val_sep:";"`
}

func TestURLValuesFromStruct(t *testing.T) {
	t.Parallel()

	s := testStruct1{
		Name:       "test",
		ID:         1,
		Populated:  true,
		Omitted:    "omitted_value",
		MultiSlice: []string{"one", "two", "three"},
		MultiArray: [3]string{"one", "two", "three"},
		SepSlice:   []string{"one", "two", "three"},
	}

	values, err := hc.URLValuesFromStruct(s)
	if err != nil {
		t.Fatalf("failed to get values for struct: %v", err)
	}

	if name := values.Get("name"); name != "test" {
		t.Errorf("'name' was '%s', expecting 'test'", name)
	} else if id := values.Get("id"); id != "1" {
		t.Errorf("'id' was '%s', expecting '1'", id)
	} else if populated := values.Get("populated"); populated != "true" {
		t.Errorf("'populated' was '%s', expecting 'true'", populated)
	} else if ms := values["multi_slice"]; len(ms) != 3 {
		t.Errorf("'multi_slice' had %d elements, expecting 3", len(ms))
	} else if ms[0] != "one" || ms[1] != "two" || ms[2] != "three" {
		t.Errorf("'multi_slice' was %v, expecting 'one', 'two', 'three'", ms)
	} else if ma := values["multi_array"]; len(ma) != 3 {
		t.Errorf("'multi_array' had %d elements, expecting 3", len(ma))
	} else if ma[0] != "one" || ma[1] != "two" || ma[2] != "three" {
		t.Errorf("'multi_slice' was %v, expecting 'one', 'two', 'three'", ms)
	} else if ss := values["sep_slice"]; len(ss) > 1 {
		t.Errorf("'sep_slice' had %d elements, expecting 1", len(ss))
	} else if ss[0] != "one;two;three" {
		t.Errorf("'sep_slice' had value '%s', expecting 'one;two;three'", ss[0])
	} else if len(values) > 6 {
		t.Errorf("received an omitted field... %+v", values)
	}
}
