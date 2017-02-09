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
			"push_int 50 halt",
			[]Item{
				{Type: ItemIdentifier, Value: "push_int", Line: 1, Pos: 0},
				{Type: ItemNumLit, Value: "50", Line: 1, Pos: 9},
				{Type: ItemIdentifier, Value: "halt", Line: 1, Pos: 12},
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
