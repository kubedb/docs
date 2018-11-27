package version

import (
	"fmt"
	"testing"
)

func TestVersion_MarshalJSON(t *testing.T) {
	v_1_2, _ := NewVersion("1.2.0-rc.1+xyz")
	v_1, _ := NewVersion("1")

	cases := []struct {
		v    *Version
		json string
		err  bool
	}{
		{v_1_2, `"1.2.0-rc.1+xyz"`, false},
		{v_1, `"1.0.0"`, false},
	}

	for _, tc := range cases {
		b, err := tc.v.MarshalJSON()
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.v)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.v, err)
		}

		if string(b) != tc.json {
			t.Fatalf("input: %s\nexpected json: %s\nactual: %s",
				tc.v, tc.json, string(b))
		}
	}
}

func TestVersion_UnmarshalJSON(t *testing.T) {
	v_1_2, _ := NewVersion("1.2.0-rc.1+xyz")
	v_1, _ := NewVersion("1")

	cases := []struct {
		v    *Version
		json string
		err  bool
	}{
		{v_1_2, `"1.2.0-rc.1+xyz"`, false},
		{v_1, `"1.0.0"`, false},
	}

	for _, tc := range cases {
		var in Version
		err := in.UnmarshalJSON([]byte(tc.json))
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.v)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.v, err)
		}
		if !tc.v.Equal(&in) {
			t.Fatalf("input: %v\nexpected: %v\nactual: %v",
				tc.v, tc.v, in)
		}
	}
}

func TestConstraints_MarshalJSON(t *testing.T) {
	v_1_2, _ := NewConstraint(">= 1.2")
	v_1, _ := NewConstraint("1.0")

	cases := []struct {
		v    Constraints
		json string
		err  bool
	}{
		{v_1_2, `">= 1.2"`, false},
		{v_1, `"1.0"`, false},
	}

	for _, tc := range cases {
		b, err := tc.v.MarshalJSON()
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.v)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.v, err)
		}

		if string(b) != tc.json {
			fmt.Println([]byte(tc.json))
			fmt.Println(b)

			t.Fatalf("input: %s\nexpected json: %s\nactual: %s",
				tc.v, tc.json, string(b))
		}
	}
}

func TestConstraints_UnmarshalJSON(t *testing.T) {
	v_1_2, _ := NewConstraint(">= 1.2")
	v_1, _ := NewConstraint("1.0.0")

	cases := []struct {
		v    Constraints
		json string
		err  bool
	}{
		{v_1_2, `">= 1.2"`, false},
		{v_1, `"1.0.0"`, false},
	}

	for _, tc := range cases {
		var in Constraints
		err := in.UnmarshalJSON([]byte(tc.json))
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.v)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.v, err)
		}
		if tc.v.String() != in.String() {
			t.Fatalf("input: %s\nexpected: %s\nactual: %s",
				tc.v, tc.v, in)
		}
	}
}
