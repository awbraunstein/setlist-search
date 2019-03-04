package searcher

import (
	"reflect"
	"testing"
)

func TestParseSetlist(t *testing.T) {
	tests := []struct {
		name     string
		setlist  string
		expected *Setlist
		err      bool
	}{
		{
			name:    "valid setlist",
			setlist: "ID{1}SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}",
			expected: &Setlist{
				showId: "1",
				sets: []Set{
					{songs: []string{"a", "b", "c", "d"}},
					{songs: []string{"x", "y", "z"}},
				},
				encore: Set{songs: []string{"aa", "bb"}},
			},
			err: false,
		},
		{
			name:     "missing ID",
			setlist:  "SET1{a,b,c,d}SET2{x,y,z}ENCORE{aa,bb}",
			expected: nil,
			err:      true,
		},
		{
			name:     "missing sets",
			setlist:  "ID{1}",
			expected: nil,
			err:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSetlist(tc.setlist)
			if tc.err && err == nil {
				t.Fatalf("expected error, but got nil")
			}
			if !tc.err && err != nil {
				t.Fatalf("got err=%v but expected nil", err)
			}
			if err != nil && !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("got: %v\nexpected: %v", got, tc.expected)
			}

		})

	}
}
