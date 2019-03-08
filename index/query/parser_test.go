package query

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  Statement
		err   bool
	}{
		{
			query: "a",
			want:  &Expression{Value: "a"},
			err:   false,
		}, {
			query: "NOT a",
			want:  &NotStatement{S: &Expression{Value: "a"}},
			err:   false,
		}, {
			query: "a AND b",
			want:  &AndStatement{Left: &Expression{Value: "a"}, Right: &Expression{Value: "b"}},
			err:   false,
		}, {
			query: "a OR b",
			want:  &OrStatement{Left: &Expression{Value: "a"}, Right: &Expression{Value: "b"}},
			err:   false,
		}, {
			query: "a OR NOT b",
			want: &OrStatement{
				Left:  &Expression{Value: "a"},
				Right: &NotStatement{S: &Expression{Value: "b"}},
			},
			err: false,
		}, {
			query: "a OR (b AND NOT c)",
			want: &OrStatement{
				Left: &Expression{Value: "a"},
				Right: &AndStatement{
					Left:  &Expression{Value: "b"},
					Right: &NotStatement{S: &Expression{Value: "c"}},
				},
			},
			err: false,
		}, {
			query: "(a",
			want:  nil,
			err:   true,
		}, {
			query: "a)",
			want:  nil,
			err:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("Query: "+tc.query, func(t *testing.T) {
			r := strings.NewReader(tc.query)
			p := NewParser(r)
			got, err := p.Parse()
			if err != nil && tc.err == false {
				t.Fatalf("Got unexpected error: %v", err)
			}
			if err == nil && tc.err == true {
				t.Fatal("Expected error, but got none")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Expected:\n%v\ngot:\n%v", tc.want, got)
			}
		})
	}
}
