package asm

import (
	"reflect"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	for i, tt := range []struct {
		input    string
		expected []Item
	}{
		{
			"push_int64 50 push_int64 -1024 halt",
			[]Item{
				{Type: ItemIdentifier, Value: "push_int64", Line: 1, Pos: 1},
				{Type: ItemNumLit, Value: "50", Line: 1, Pos: 12},
				{Type: ItemIdentifier, Value: "push_int64", Line: 1, Pos: 15},
				{Type: ItemNumLit, Value: "-1024", Line: 1, Pos: 26},
				{Type: ItemIdentifier, Value: "halt", Line: 1, Pos: 32},
			},
		},
	} {
		lex := NewLexer(strings.NewReader(tt.input))
		var items []Item
		for lex.Scanning() {
			items = append(items, lex.Item())
		}
		if !reflect.DeepEqual(items, tt.expected) {
			t.Errorf("%d. expecting: %#v\nreceived: %#v", i, tt.expected, items)
		}
	}
}
