package asm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/bruston/lil/bytecode"
)

type Parser struct {
	lex          *Lexer
	instructions []instruction
	out          *bytes.Buffer
	labels       map[string]int
	vars         map[string]int
}

const (
	pseudoInstructionLabel = bytecode.OpLast + 1
	pseudoInstructionVar   = bytecode.OpLast + 2
)

type instruction struct {
	op   byte
	arg  interface{}
	line int
	pos  int
}

func (p *Parser) Parse() error {
	for p.lex.Scanning() {
		itm := p.lex.Item()
		ins := instruction{}
		switch itm.Type {
		case ItemLabel:
			itm.Value = itm.Value[1:]
			p.instructions = append(p.instructions, instruction{pseudoInstructionLabel, itm.Value, itm.Line, itm.Pos})
			continue
		case ItemVar:
			p.lex.scan()
			itm := p.lex.Item()
			if itm.Type != ItemIdentifier {
				return fmt.Errorf("expecting identifier at line %d pos %d", itm.Line, itm.Pos)
			}
			if _, ok := p.vars[itm.Value]; ok {
				return fmt.Errorf("variable %d already declared at line %d pos %d", itm.Line, itm.Pos)
			}
			p.vars[itm.Value] = len(p.vars)
			continue
		}
		op, ok := imap[itm.Value]
		if itm.Type != ItemIdentifier || !ok {
			return fmt.Errorf("invalid instruction at line %d pos %d", itm.Line, itm.Pos)
		}
		ins.op = op
		ins.line, ins.pos = itm.Line, itm.Pos
		switch ins.op {
		case bytecode.OpCall, bytecode.OpJump, bytecode.OpJumpEq, bytecode.OpJumpGT, bytecode.OpJumpLT,
			bytecode.OpJumpTrue, bytecode.OpJumpFalse, bytecode.OpJumpNotEq:
			p.lex.scan()
			arg := p.lex.Item()
			if arg.Type != ItemIdentifier {
				return fmt.Errorf("expecting label identifer at line %d pos %d", itm.Line, itm.Pos)
			}
			ins.arg = arg.Value
			p.instructions = append(p.instructions, ins)
			continue
		case bytecode.OpLoad, bytecode.OpStore:
			p.lex.scan()
			arg := p.lex.Item()
			if arg.Type != ItemIdentifier {
				return fmt.Errorf("expecting label identifer at line %d pos %d", itm.Line, itm.Pos)
			}
			ins.arg = arg.Value
			p.instructions = append(p.instructions, ins)
			continue
		case bytecode.OpPushUint8:
			p.lex.scan()
			arg := p.lex.Item()
			if arg.Type != ItemNumLit {
				fmt.Errorf("expecting numeric argument, got type: %d at line %d pos %d", arg.Type, itm.Line, itm.Pos)
			}
			n, err := strconv.ParseUint(arg.Value, 10, 64)
			if err != nil || n > 255 {
				return fmt.Errorf("invalid uint8 at line %d pos %d", itm.Line, itm.Pos)
			}
			ins.arg = uint8(n)
			p.instructions = append(p.instructions, ins)
		case bytecode.OpPushInt64:
			p.lex.scan()
			arg := p.lex.Item()
			if arg.Type != ItemNumLit {
				return fmt.Errorf("invalid int64 at line %d pos %d", itm.Line, itm.Pos)
			}
			n, err := strconv.ParseInt(arg.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int64 at line %d pos %d", itm.Line, itm.Pos)
			}
			ins.arg = n
			p.instructions = append(p.instructions, ins)
		default:
			p.instructions = append(p.instructions, ins)
		}
	}
	return nil
}

func NewParser(l *Lexer) *Parser {
	return &Parser{
		lex:    l,
		out:    &bytes.Buffer{},
		labels: make(map[string]int),
		vars:   make(map[string]int),
	}
}

func (p *Parser) Compile() ([]byte, int, int, error) {
	var pos int
	for _, v := range p.instructions {
		if v.op == bytecode.OpLast+1 {
			p.labels[v.arg.(string)] = pos
			continue
		}
		if v.op == bytecode.OpStore || v.op == bytecode.OpLoad {
			i, ok := p.vars[v.arg.(string)]
			if !ok {
				return nil, 0, 0, errors.New("no such var: " + v.arg.(string))
			}
			v.arg = int64(i)
		}
		if isJump(v.op) || v.op == bytecode.OpCall {
			label, ok := v.arg.(string)
			if !ok {
				return nil, 0, 0, errors.New("no label specified")
			}
			n, ok := p.labels[label]
			if !ok {
				return nil, 0, 0, fmt.Errorf("no such label: %s", label)
			}
			v.arg = int64(n)
		}
		n, err := bytecode.Encode(p.out, v.op, v.arg)
		if err != nil {
			return nil, 0, 0, err
		}
		pos += n
	}
	if start, ok := p.labels["main"]; ok {
		return p.out.Bytes(), start, len(p.vars), nil
	}
	return p.out.Bytes(), 0, len(p.vars), nil
}

func isJump(op byte) bool {
	return op == bytecode.OpJump || op == bytecode.OpJumpEq || op == bytecode.OpJumpTrue || op == bytecode.OpJumpFalse ||
		op == bytecode.OpJumpLT || op == bytecode.OpJumpGT || op == bytecode.OpJumpNotEq
}

var imap = map[string]byte{
	"nop":        bytecode.OpNOP,
	"halt":       bytecode.OpHalt,
	"push_int64": bytecode.OpPushInt64,
	"push_uint8": bytecode.OpPushUint8,
	"push_zero":  bytecode.OpPushZero,
	"push_one":   bytecode.OpPushOne,
	"store":      bytecode.OpStore,
	"load":       bytecode.OpLoad,
	"to_int64":   bytecode.OpToInt64,
	"to_uint8":   bytecode.OpToUint8,
	"print":      bytecode.OpPrint,
	"print_ch":   bytecode.OpPrintCh,
	"drop":       bytecode.OpDrop,
	"dup":        bytecode.OpDup,
	"swap":       bytecode.OpSwap,
	"jump":       bytecode.OpJump,
	"jump_true":  bytecode.OpJumpTrue,
	"jump_false": bytecode.OpJumpTrue,
	"jump_eq":    bytecode.OpJumpEq,
	"jump_ne":    bytecode.OpJumpNotEq,
	"jump_lt":    bytecode.OpJumpLT,
	"jump_gt":    bytecode.OpJumpGT,
	"add":        bytecode.OpAdd,
	"sub":        bytecode.OpSub,
	"mul":        bytecode.OpMul,
	"div":        bytecode.OpDiv,
	"inc":        bytecode.OpInc,
	"dec":        bytecode.OpDec,
	"mod":        bytecode.OpMod,
	"and":        bytecode.OpAnd,
	"or":         bytecode.OpOr,
	"xor":        bytecode.OpXOR,
	"not":        bytecode.OpNot,
	"call":       bytecode.OpCall,
	"ret":        bytecode.OpRet,
}

func Compile(src io.Reader, dst io.Writer) error {
	parser := NewParser(NewLexer(src))
	if err := parser.Parse(); err != nil {
		return err
	}
	code, start, dataElements, err := parser.Compile()
	if err != nil {
		return err
	}
	if _, err := bytecode.WriteHeader(dst, start, dataElements); err != nil {
		return err
	}
	_, err = dst.Write(code)
	return err
}
