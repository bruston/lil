package bytecode

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	OpNOP byte = iota
	OpHalt
	OpPrint
	OpPrintCh
	OpPushUint8
	OpPushInt64
	OpToInt64
	OpToUint8
	OpPushZero
	OpPushOne
	OpStore
	OpLoad
	OpCreateArray
	OpArrayLoad
	OpArrayStore
	OpDrop
	OpDup
	OpSwap
	OpJump
	OpJumpTrue
	OpJumpFalse
	OpJumpEq
	OpJumpNotEq
	OpJumpLT
	OpJumpGT
	OpMov
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpInc
	OpDec
	OpMod
	OpAnd
	OpOr
	OpXOR
	OpNot
	OpCall
	OpRet
	OpLast // Keep this as the final code in the list.
)

func WriteHeader(w io.Writer, start, dataElements int) (int, error) {
	var written int
	for _, v := range []int{start, dataElements} {
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutVarint(buf, int64(v))
		if _, err := w.Write(buf[:size]); err != nil {
			return 0, err
		}
		written += size
	}
	return written, nil
}

var ErrInvalidArgument = errors.New("invalid argument")

func Encode(w io.Writer, op byte, arg interface{}) (int, error) {
	ins, ok := imap[op]
	if !ok {
		return 0, fmt.Errorf("no such op code: %d", op)
	}
	switch ins.arg {
	case argNone:
		if arg != nil {
			return 0, fmt.Errorf("op code %s has no arguments but one was specified", ins.name)
		}
		return w.Write([]byte{op})
	case argInt:
		n, ok := arg.(int64)
		if !ok {
			return 0, ErrInvalidArgument
		}
		buf := make([]byte, binary.MaxVarintLen64)
		size := binary.PutVarint(buf, n)
		b := make([]byte, 0, size+1)
		b = append(b, op)
		b = append(b, buf[:size]...)
		return w.Write(b)
	case argUint:
		n, ok := arg.(uint8)
		if !ok {
			return 0, ErrInvalidArgument
		}
		return w.Write([]byte{op, byte(n)})
	}
	// should be unreachable
	return 0, fmt.Errorf("invalid instruction argument type: %d", ins.arg)
}

const (
	argNone = iota
	argInt
	argUint
)

type instruction struct {
	name string
	arg  int
}

var imap = map[byte]instruction{
	OpNOP:       {"nop", argNone},
	OpHalt:      {"halt", argNone},
	OpPushInt64: {"push_int64", argInt},
	OpPushUint8: {"push_uint8", argUint},
	OpPushZero:  {"push_zero", argNone},
	OpPushOne:   {"push_one", argNone},
	OpStore:     {"store", argInt},
	OpLoad:      {"load", argInt},
	OpToInt64:   {"to_int64", argNone},
	OpToUint8:   {"to_uint8", argNone},
	OpPrint:     {"print", argNone},
	OpPrintCh:   {"print_ch", argNone},
	OpDrop:      {"drop", argNone},
	OpDup:       {"dup", argNone},
	OpSwap:      {"swap", argNone},
	OpJump:      {"jump", argInt},
	OpJumpTrue:  {"jump_true", argInt},
	OpJumpFalse: {"jump_false", argInt},
	OpJumpEq:    {"jump_eq", argInt},
	OpJumpNotEq: {"jump_ne", argInt},
	OpJumpLT:    {"jump_lt", argInt},
	OpJumpGT:    {"jump_gt", argInt},
	OpAdd:       {"add", argNone},
	OpSub:       {"sub", argNone},
	OpMul:       {"mul", argNone},
	OpDiv:       {"div", argNone},
	OpInc:       {"inc", argNone},
	OpDec:       {"dec", argNone},
	OpMod:       {"mod", argNone},
	OpAnd:       {"and", argNone},
	OpOr:        {"or", argNone},
	OpXOR:       {"xor", argNone},
	OpNot:       {"not", argNone},
	OpCall:      {"call", argInt},
	OpRet:       {"ret", argNone},
}
