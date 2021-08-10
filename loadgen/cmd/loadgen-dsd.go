package main

import (
	"bufio"
	"os"

	"github.com/djmitche/tagset/loadgen"
)

// On a macbook, this is capable of about 2 million lines / second

func main() {
	tlg := loadgen.DSDTagLineGenerator()

	out := bufio.NewWriter(os.Stdout)

	for l := range tlg.GetLines() {
		_, err := out.Write(l)
		if err != nil {
			panic(err)
		}
		_, err = out.WriteString("\n")
		if err != nil {
			panic(err)
		}
	}
}
