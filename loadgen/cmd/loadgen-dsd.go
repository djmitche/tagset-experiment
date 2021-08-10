package main

import (
	"fmt"

	"github.com/djmitche/tagset/loadgen"
)

func main() {
	tlg := loadgen.DSDTagLineGenerator()

	for l := range tlg.GetLines() {
		fmt.Printf("%s\n", l)
	}
}
