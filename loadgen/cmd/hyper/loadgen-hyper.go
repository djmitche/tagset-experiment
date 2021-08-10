package main

import "os"

// This producer tries to write bogus tag lines as fast as possible.  On a
// macbook, this is capable of about 700MB/s, or a line every 175ns

func main() {
	const linesize = 128
	line := make([]byte, linesize)

	for i := range line {
		line[i] = 'a'
	}
	line[len(line)-1] = '\n'

	const bufsize = 64 * linesize * 32
	buf := []byte{}
	for i := 0; i < bufsize; i += linesize {
		buf = append(buf, line...)
	}

	for {
		n, err := os.Stdout.Write(buf)
		if err != nil {
			panic(err)
		}
		if n != len(buf) {
			panic(err)
		}
	}
}
