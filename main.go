package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bruston/lil/asm"
	"github.com/bruston/lil/vm"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stdout, "Usage is:\nlil run file.lil\nlil asm file.asm\n")
		os.Exit(0)
	}
	cmd := os.Args[1]
	switch cmd {
	case "run":
		m, err := vm.Open(os.Args[2])
		if err != nil {
			fmt.Println(os.Stderr, "error opening vm image:", err)
			os.Exit(1)
		}
		if err := m.Exec(); err != nil {
			fmt.Fprintln(os.Stderr, "error encountered during execution:", err)
			os.Exit(1)
		}
	case "asm":
		var outPath string
		if len(os.Args) < 4 {
			outPath = "out.lil"
		} else {
			outPath = os.Args[3]
		}
		f, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		out, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to create output file:", err)
			os.Exit(1)
		}
		if err := asm.Compile(f, out); err != nil {
			fmt.Fprintln(os.Stderr, "error compiling asm:", err)
			os.Exit(1)
		}
		if err := out.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "error closing output file, contents may not have been written correctly:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command, valid commands are asm and run")
		os.Exit(1)
	}
}
