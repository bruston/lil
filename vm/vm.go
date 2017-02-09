package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/bruston/lil/bytecode"
)

const (
	DefaultStackSize     = 1024
	DefaultCallStackSize = 1024
)

type Stack struct {
	top      int
	elements []Value
}

func (s *Stack) Push(v Value) {
	s.top++
	s.elements[s.top] = v
}

func (s *Stack) Pop() Value {
	v := s.elements[s.top]
	s.elements[s.top] = nil
	s.top--
	return v
}

func (s *Stack) Peek() Value {
	return s.elements[s.top]
}

func (s *Stack) Swap() {
	s.elements[s.top], s.elements[s.top-1] = s.elements[s.top-1], s.elements[s.top]
}

func (s *Stack) Dup() {
	v := s.elements[s.top]
	s.top++
	s.elements[s.top] = v
}

func NewStack(size int) *Stack {
	return &Stack{
		top:      -1,
		elements: make([]Value, size),
	}
}

type Machine struct {
	Instructions []byte
	Stack        *Stack
	CallStack    *Stack
	Data         []Value
	IP           int
	SP           int
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
}

func NewMachine(stackSize, callStackSize int) *Machine {
	return &Machine{
		Stack:     NewStack(stackSize),
		CallStack: NewStack(callStackSize),
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}
}

var ErrInvalidVarint = errors.New("supplied varint is invalid")

func (m *Machine) readVarint() (int64, error) {
	n, read := binary.Varint(m.Instructions[m.IP:])
	if read <= 0 {
		return 0, ErrInvalidVarint
	}
	m.IP += read - 1
	return n, nil
}

func (m *Machine) Exec() error {
	if len(m.Instructions) == 0 {
		return nil
	}
	for {
		switch m.Instructions[m.IP] {
		case bytecode.OpPushZero:
			m.Stack.Push(Int64{ValueInt64, 0})
		case bytecode.OpPushOne:
			m.Stack.Push(Int64{ValueInt64, 1})
		case bytecode.OpPushUint8:
			m.IP++
			m.Stack.Push(Uint8{ValueUint8, m.Instructions[m.IP]})
		case bytecode.OpPushInt64:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			m.Stack.Push(Int64{ValueInt64, n})
		case bytecode.OpPrint:
			v := m.Stack.Pop()
			fmt.Fprint(os.Stdout, v)
		case bytecode.OpPrintCh:
			if m.Stack.Peek().Type() != ValueUint8 {
				return errors.New("expecting Uint8 arg for PrintCh")
			}
			v := m.Stack.Pop()
			fmt.Fprint(os.Stdout, string(v.Value().(uint8)))
		case bytecode.OpDrop:
			m.Stack.Pop()
		case bytecode.OpStore:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			m.Data[int(n)] = m.Stack.Pop()
		case bytecode.OpLoad:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			m.Stack.Push(m.Data[int(n)])
		case bytecode.OpToInt64:
			if m.Stack.Peek().Type() != ValueUint8 {
				errors.New("cannot convert non-uint8 value to int64")
			}
			v := m.Stack.Pop()
			m.Stack.Push(Int64{ValueInt64, int64(v.Value().(uint8))})
		case bytecode.OpToUint8:
			if m.Stack.Peek().Type() != ValueInt64 {
				return errors.New("cannot convert non-int64 value to uint8")
			}
			v := m.Stack.Pop()
			if v.Value().(int64) < 0 || v.Value().(int64) > 255 {
				return errors.New("unable to convert int64 to uint8: outside of range: 0-255")
			}
			m.Stack.Push(Uint8{ValueUint8, uint8(v.Value().(int64))})
		case bytecode.OpAdd:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() != ValueInt64 || b.Type() != ValueInt64 {
				return errors.New("attempted addition on non-int64 values")
			}
			m.Stack.Push(Int64{ValueInt64, a.Value().(int64) + b.Value().(int64)})
		case bytecode.OpSub:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() != ValueInt64 || b.Type() != ValueInt64 {
				return errors.New("attempted subtraction on non-int64 values")
			}
			m.Stack.Push(Int64{ValueInt64, a.Value().(int64) - b.Value().(int64)})
		case bytecode.OpMul:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() != ValueInt64 || b.Type() != ValueInt64 {
				return errors.New("attempted multiplication on non-int64 values")
			}
			m.Stack.Push(Int64{ValueInt64, a.Value().(int64) * b.Value().(int64)})
		case bytecode.OpMod:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() != ValueInt64 || b.Type() != ValueInt64 {
				return errors.New("attempted mod on non-int64 values")
			}
			m.Stack.Push(Int64{ValueInt64, a.Value().(int64) % b.Value().(int64)})
		case bytecode.OpSwap:
			m.Stack.Swap()
		case bytecode.OpDup:
			m.Stack.Dup()
		case bytecode.OpInc:
			switch v := m.Stack.Pop().(type) {
			case Int64:
				v.Val++
				m.Stack.Push(v)
			case Uint8:
				v.Val++
				m.Stack.Push(v)
			default:
				return errors.New("attempted to increment a non-numeric type")
			}
		case bytecode.OpDec:
			switch v := m.Stack.Pop().(type) {
			case Int64:
				v.Val--
				m.Stack.Push(v)
			case Uint8:
				v.Val--
				m.Stack.Push(v)
			default:
				return errors.New("attempted to decrement a non-numeric type")
			}
		case bytecode.OpJump:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			m.IP = int(n) - 1
		case bytecode.OpJumpTrue:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			v := m.Stack.Pop()
			if v.Value() != 0 {
				m.IP = int(n) - 1
			}
		case bytecode.OpJumpFalse:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			v := m.Stack.Pop()
			if v.Value() == 0 {
				m.IP = int(n) - 1
			}
		case bytecode.OpJumpEq:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Value() == b.Value() {
				m.IP = int(n) - 1
			}
		case bytecode.OpJumpNotEq:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Value() != b.Value() {
				m.IP = int(n) - 1
			}
		case bytecode.OpJumpLT:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			b, a := m.Stack.Pop(), m.Stack.Pop()
			ac, ok := a.(Comparable)
			if !ok {
				return errors.New("attempting to compare incomparable types")
			}
			if ac.Compare(b) == -1 {
				m.IP = int(n) - 1
			}
		case bytecode.OpJumpGT:
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			b, a := m.Stack.Pop(), m.Stack.Pop()
			ac, ok := a.(Comparable)
			if !ok {
				return errors.New("attempting to compare incomparable types")
			}
			if ac.Compare(b) == 1 {
				m.IP = int(n) - 1
			}
		case bytecode.OpOr:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() == ValueInt64 && b.Type() == ValueInt64 {
				m.Stack.Push(Int64{ValueInt64, a.Value().(int64) | b.Value().(int64)})
			} else if a.Type() == ValueUint8 && b.Type() == ValueUint8 {
				m.Stack.Push(Uint8{ValueUint8, a.Value().(uint8) | b.Value().(uint8)})
			} else {
				return errors.New("attempting bitwise OR on different types")
			}
		case bytecode.OpAnd:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() == ValueInt64 && b.Type() == ValueInt64 {
				m.Stack.Push(Int64{ValueInt64, a.Value().(int64) & b.Value().(int64)})
			} else if a.Type() == ValueUint8 && b.Type() == ValueUint8 {
				m.Stack.Push(Uint8{ValueUint8, a.Value().(uint8) & b.Value().(uint8)})
			} else {
				return errors.New("attempting bitwise AND on incompatible types")
			}
		case bytecode.OpXOR:
			b, a := m.Stack.Pop(), m.Stack.Pop()
			if a.Type() == ValueInt64 && b.Type() == ValueInt64 {
				m.Stack.Push(Int64{ValueInt64, a.Value().(int64) ^ b.Value().(int64)})
			} else if a.Type() == ValueUint8 && b.Type() == ValueUint8 {
				m.Stack.Push(Uint8{ValueUint8, a.Value().(uint8) ^ b.Value().(uint8)})
			} else {
				return errors.New("attempting bitwise XOR on incompatible types")
			}
		case bytecode.OpCall:
			m.CallStack.Push(Int64{ValueInt64, int64(m.IP)})
			m.IP++
			n, err := m.readVarint()
			if err != nil {
				return err
			}
			m.IP = int(n) - 1
		case bytecode.OpRet:
			m.IP = int(m.CallStack.Pop().Value().(int64))
		case bytecode.OpNOP:
		case bytecode.OpHalt:
			return nil
		}
		m.IP++
	}
	return nil
}

func Open(path string) (*Machine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := NewMachine(DefaultStackSize, DefaultCallStackSize)
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	n, size := binary.Varint(b)
	m.IP = int(n)
	m.Instructions = b[size:]
	n, size = binary.Varint(m.Instructions)
	m.Instructions = m.Instructions[size:]
	m.Data = make([]Value, n)
	return m, nil
}
