package vm

import (
	"fmt"
	"reflect"
)

type ValueType byte

const (
	ValueInt64 ValueType = iota
	ValueUint8
	ValueArray
	ValuePair
)

func (vt ValueType) Type() ValueType { return vt }

type Comparable interface {
	Compare(Value) int
}

type Value interface {
	Type() ValueType
	Value() interface{}
}

type Appendable interface {
	Append(Value)
}

type Int64 struct {
	ValueType
	Val int64
}

func (i Int64) Value() interface{} { return i.Val }

func (i Int64) String() string { return fmt.Sprintf("%d", i.Val) }

func (i Int64) Equal(v Value) bool {
	switch n := v.Value().(type) {
	case Int64:
		return n.Value() == i.Val
	case Uint8:
		return int64(n.Value().(uint8)) == i.Val
	}
	return false
}

func (i Int64) Compare(v Value) int {
	switch n := v.(type) {
	case Int64:
		if i.Val < n.Val {
			return -1
		}
		if i.Val == n.Val {
			return 0
		}
		return 1
	case Uint8:
		if i.Val < int64(n.Val) {
			return -1
		}
		if i.Val == int64(n.Val) {
			return 0
		}
		return 1
	}
	return -2 // should never happen
}

type Uint64 struct {
	ValueType
	Val uint64
}

type Uint8 struct {
	ValueType
	Val uint8
}

func (u Uint8) Value() interface{} { return u.Val }

func (u Uint8) String() string { return fmt.Sprintf("%d", u.Val) }

func (u Uint8) Equal(v Value) bool {
	switch v := v.(type) {
	case Int64:
		if v.Value().(int64) > 255 {
			return false
		}
		return uint8(v.Value().(int64)) == u.Val
	case Uint8:
		return v.Value().(uint8) == u.Val
	}
	return false
}

func (u Uint8) Compare(v Value) int { return 0 }

type Array struct {
	ValueType
	elements []Value
}

func (a *Array) Value() interface{} { return a.elements }

func (a *Array) Append(v Value) { a.elements = append(a.elements, v) }

func (a *Array) Index(i int) Value { return a.elements[i] }

func (a *Array) Set(i int, v Value) { a.elements[i] = v }

func (a *Array) Len() int { return len(a.elements) }

func (a *Array) Cap() int { return cap(a.elements) }

func (a *Array) Equal(v Value) bool {
	if v.Type() == ValueArray {
		return reflect.DeepEqual(v.Value(), a.elements)
	}
	return false
}

type Pair struct {
	ValueType
	Elements [2]Value
}

func (p Pair) Value() interface{} { return p.Elements }
